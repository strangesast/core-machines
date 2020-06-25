package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const timestampFormat = "2006-01-02T15:04:05.999Z"

var (
	d  net.Dialer
	wg sync.WaitGroup
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// kafka stuff
	kafkaVersion, err := sarama.ParseKafkaVersion(getEnv("KAFKA_VERSION", "2.4.1"))
	if err != nil {
		log.Fatalln("failed to parse kafka version string")
	}

	config := sarama.NewConfig()
	config.Net.DialTimeout = 5 * time.Minute
	config.Version = kafkaVersion
	config.Producer.Return.Successes = true
	config.ClientID = "golang-serial-monitoring-producer"

	kafkaHosts := strings.Split(getEnv("KAFKA_HOSTS", "localhost:9092"), ",")
	log.Printf("using KAFKA_VERSION='%v', KAFKA_HOSTS='%v'\n", kafkaVersion, kafkaHosts)
	kafkaClient, err := sarama.NewClient(kafkaHosts, config)
	if err != nil {
		log.Fatalf("failed to create sarama kafka client: %v", err)
	}

	producer, err := sarama.NewAsyncProducerFromClient(kafkaClient)
	if err != nil {
		log.Fatalf("failed to create sarama producer: %v", err)
	}

	machineID := getEnv("MACHINE_ID", "unknown")
	log.Printf("using MACHINE_ID='%s'\n", machineID)

	// tcp socket stuff
	adapterHost := getEnv("ADAPTER_HOST", "localhost:7878")
	log.Printf("using ADAPTER_HOST='%s'", adapterHost)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	// if interrupted, cancel
	go func() {
		select {
		case <-signals:
			fmt.Println("cancelling...")
			cancel()
		case <-ctx.Done():
		}
	}()

	conn, err := d.DialContext(context.Background(), "tcp", adapterHost)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	ticker := time.NewTicker(10 * time.Second) // mtconnect requires PING at some frequency

	lines := make(chan string, 100) // buffer a few lines

	var cnt int

	// send heartbeat ping
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(conn, "* PING")
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// read from successes (unnecessary)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-producer.Successes():
				cnt++
			case <-ctx.Done():
				return
			}
		}
	}()

	// read from lines channel, send to kafka
	wg.Add(1)
	go func() {
		defer wg.Done()
		var message *sarama.ProducerMessage
		for {
			select {
			case <-ctx.Done():
				producer.AsyncClose()
				return
			case line := <-lines:
				message = &sarama.ProducerMessage{
					Topic: "input-text",
					Value: sarama.StringEncoder(line),
					Key:   sarama.StringEncoder(machineID),
				}
				producer.Input() <- message
			case err := <-producer.Errors():
				log.Printf("err: %v", err)
			}
		}
	}()

	// wait for ctx to close, and close connection
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		conn.Close()
	}()

	// read lines from connection, send to channel
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()

			// may want to pass this to kafka
			if strings.HasPrefix(line, "* PONG") {
				continue
			}

			values := strings.Split(line, "|")
			timestampString := values[0]

			_, err = time.Parse(timestampFormat, timestampString)
			if err != nil {
				log.Printf("failed to parse timestamp: %s (%s)\n", timestampString, timestampFormat)
				continue
			}

			lines <- line
		}

		if ctx.Err() != nil {
			return
		} else if err := scanner.Err(); err != nil {
			log.Fatalf("failed to read from connection: %v", err)
			cancel()
		}
	}()

	wg.Wait()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

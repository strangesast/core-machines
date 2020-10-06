package main

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type consumerHandler struct{}

func (consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// splitter := regexp.MustCompile(`(?<!\\)[|]`)
	// "2020-08-02T12:53:57.909089Z"
	format := "2006-01-02T15:04:05.000000Z"

	for msg := range claim.Messages() {
		parts := strings.Split(string(msg.Value), "|")
		if l := len(parts); l > 0 {
			ts, err := time.Parse(format, parts[0])
			if err != nil {
				log.Fatal(err)
			}
			// log.Printf("topic %q. partition %d offset %d\n%v\n%s\n", msg.Topic, msg.Partition, msg.Offset, ts, strings.Join(parts[1:], "|"))
			if strings.Contains(string(msg.Value), "part_count") {
				log.Printf("datetime %v, %s\n", ts, string(msg.Value))
			}
		}
	}
	return nil
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// kafka stuff
	//kafkaVersion, err := sarama.ParseKafkaVersion(getEnv("KAFKA_VERSION", "2.4.1"))
	kafkaVersion, err := sarama.ParseKafkaVersion("2.4.1")
	if err != nil {
		log.Fatalln("failed to parse kafka version string")
	}

	config := sarama.NewConfig()
	config.Net.DialTimeout = 5 * time.Minute
	config.Version = kafkaVersion
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.ClientID = "golang-machines-consumer"

	kafkaHosts := strings.Split("localhost:9092", ",")

	log.Printf("using KAFKA_VERSION='%v', KAFKA_HOSTS='%v'\n", kafkaVersion, kafkaHosts)

	kafkaClient, err := sarama.NewClient(kafkaHosts, config)

	if err != nil {
		log.Fatalf("failed to create sarama kafka client: %v", err)
	}

	consumer, err := sarama.NewConsumerGroupFromClient("group-0", kafkaClient)

	if err != nil {
		log.Fatalln("failed to create consumer group")
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for err := range consumer.Errors() {
			log.Printf("error consuming: %v", err)
		}
	}()

	go func() {
		select {
		case val := <-signals:
			log.Printf("got signal (%v)\n", val)
			cancel()
		case <-ctx.Done():
		}
		consumer.Close()
	}()

	topics := []string{"input-text"}

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		default:
		}
		handler := consumerHandler{}
		err := consumer.Consume(ctx, topics, handler)
		if err != nil {
			log.Fatal(err)
		}
	}
}

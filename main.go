package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	ptypes "github.com/golang/protobuf/ptypes"
	pb "github.com/strangesast/core/monitoring/proto"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// const baseURL = `https://smstestbed.nist.gov/vds`

var baseURL string

func main() {
	baseURL = getEnv("BASE_URL", "http://localhost:5000")
	log.Printf(`using base url '%s'`, baseURL)

	onRegister := make(chan *Client)
	hub := newHub(onRegister)

	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	config := sarama.NewConfig()
	config.Net.DialTimeout = 5 * time.Minute

	kafkaVersion, err := sarama.ParseKafkaVersion(getEnv("KAFKA_VERSION", "2.4.1"))
	handleErr(err)
	kafkaHosts := strings.Split(getEnv("KAFKA_HOSTS", "localhost:9092"), ",")
	log.Printf(`using kafka host %s (v%s)\n`, strings.Join(kafkaHosts, " and "), kafkaVersion)

	config.Version = kafkaVersion
	config.ClientID = "golang-monitoring-producer"
	// config.Producer.Return.Successes = true

	kafkaClient, err := sarama.NewClient(kafkaHosts, config)
	handleErr(err)
	kafkaClusterAdmin, err := sarama.NewClusterAdminFromClient(kafkaClient)
	handleErr(err)
	producer, err := sarama.NewAsyncProducerFromClient(kafkaClient)
	handleErr(err)

	topics, err := kafkaClusterAdmin.ListTopics()
	handleErr(err)

	fmt.Println("topics:")
	for topicName, topicDetail := range topics {
		fmt.Printf("%s %v\n", topicName, topicDetail)
		// kafkaClusterAdmin.DeleteTopic(topicName)
	}

	go func() {
		// devices := probe()
		// if len(devices.Devices) == 0 {
		// 	fmt.Println("no devices?")
		// 	return
		// }
		// device := devices.Devices[0]

		// serve()
		//fmt.Printf("Device %s\n", device.Name)
		init := current("")

		// pprint(init)

		fmt.Println("input messsage")

		msg := &pb.SampleGroup{Samples: printStreams(init), LastSequence: int64(init.Header.LastSequence)}

		b0, err := proto.Marshal(msg)
		handleErr(err)

		kMessage := &sarama.ProducerMessage{Topic: "input", Value: sarama.ByteEncoder(b0)}
		producer.Input() <- kMessage

		marshaller := jsonpb.Marshaler{}
		b1, err := marshaller.MarshalToString(msg)
		handleErr(err)
		hub.broadcast <- []byte(b1)

		seq := init.Header.NextSequence
		fmt.Printf("seq: %v\n", seq)

		// , streams.Header.FirstSequence
		samples := sample("", seq)

		for {
			select {
			case s := <-samples:
				fmt.Println("got samples")

				msg := &pb.SampleGroup{Samples: printStreams(s), LastSequence: int64(s.Header.LastSequence)}

				b0, err := proto.Marshal(msg)
				handleErr(err)

				b1, err := marshaller.MarshalToString(msg)
				handleErr(err)

				kMessage := &sarama.ProducerMessage{Topic: "input", Value: sarama.ByteEncoder(b0)}

				producer.Input() <- kMessage
				hub.broadcast <- []byte(b1)

			case c := <-onRegister:
				fmt.Println("sending init")

				kMessage := &pb.SampleGroup{Samples: printStreams(init)}
				b, _ := marshaller.MarshalToString(kMessage)
				c.send <- []byte(b)
			}
		}
	}()

	err = http.ListenAndServe(*addr, nil)
	handleErr(err)

	// kafkaClient.Close()
}

func printStreams(streams mMTConnectStreams) []*pb.Sample {
	var samples []*pb.Sample

	for _, stream := range streams.Streams {
		for _, component := range stream.Components {
			for _, item := range component.Samples.Items {
				sample := parseSample(item, stream, component, pb.Sample_SAMPLE)
				samples = append(samples, &sample)
			}
			for _, item := range component.Events.Items {
				sample := parseSample(item, stream, component, pb.Sample_EVENT)
				samples = append(samples, &sample)
			}
			for _, item := range component.Condition.Items {
				sample := parseSample(item, stream, component, pb.Sample_CONDITION)
				samples = append(samples, &sample)
			}
		}
	}
	return samples
}

func parseSample(item mComponentSample, stream mStream, component mComponentStream, stype pb.Sample_SampleType) pb.Sample {
	attrs := make(map[string]string)

	for _, attr := range item.Attrs {
		n := attr.Name.Local
		switch n {
		case "dataItemId",
			"timestamp":
			continue
		default:
			attrs[n] = attr.Value
		}
	}
	// yeesh
	format1 := "2006-01-02T15:04:05.999999Z"
	format2 := "2006-01-02T15:04:05.999999"

	var ts time.Time
	ts, err := time.Parse(format1, item.Timestamp)

	if err != nil {
		ts, err = time.Parse(format2, item.Timestamp)
	}
	handleErr(err)

	t, err := ptypes.TimestampProto(ts)
	handleErr(err)

	return pb.Sample{
		Device:    stream.DeviceName,
		Itemid:    item.DataItemID,
		Sequence:  int64(item.Sequence),
		Component: component.ComponentID,
		Type:      stype,
		Value:     item.Value,
		Timestamp: t,
		Tag:       item.XMLName.Local,
		Attrs:     attrs,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

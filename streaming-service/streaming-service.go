package streamingservice

import (
	"context"
	"log"
	mitsuko "mitsuko-relay/lib/payloadbuilder/src/proto/pb/mitsuko/relay"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type StreamPayload struct {
	Topic   string
	Payload *mitsuko.RelayPayload
}

func Run(seeds []string, strmc chan *StreamPayload) {
	ctx := context.Background()

	opts := []kgo.Opt{}
	opts = append(opts,
		kgo.SeedBrokers(seeds...),
		kgo.ProduceRequestTimeout(time.Second*5),
	)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	for payload := range strmc {
		client.Produce(
			ctx,
			&kgo.Record{
				Topic:     payload.Topic,
				Key:       []byte(payload.Payload.Key),
				Value:     payload.Payload.Pkt,
				Timestamp: time.UnixMilli(payload.Payload.Timestamp),
			},
			producerCb,
		)
	}
}

func producerCb(r *kgo.Record, err error) {
	if err != nil {
		log.Println(err)
	}
}

package main

import (
	"context"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/hack"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"math/rand"
	"strconv"
)

func main() {
	producer := kafka.MustKafkaProducer(&kafka.KafkaProducerConf{
		Topic:   "chat-test-topic",
		Brokers: []string{"127.0.0.1:9092"},
	})

	// doSendMessage(producer)
	doSendMessageV2(producer)
}

func doSendMessage(producer *kafka.Producer) {
	i := 0

	for {
		k := "sync.TL_sync_pushUpdates#136817694#" + strconv.FormatInt(int64(i), 10)
		v := strconv.FormatInt(rand.Int63(), 10)
		_, _, err := producer.SendMessage(context.Background(), k, hack.Bytes(v))
		if err != nil {
			fmt.Println("error - ", err)
		}
		fmt.Println("k:", k, "v:", v)
		i = i + 1
		// time.Sleep(100 * time.Millisecond)
	}
}

func doSendMessageV2(producer *kafka.Producer) {
	i := 0

	for {
		k := strconv.FormatInt(int64(i%10), 10)
		v := strconv.FormatInt(rand.Int63(), 10)
		_, _, err := producer.SendMessageV2(context.Background(), "sendMessage", k, hack.Bytes(v))
		if err != nil {
			fmt.Println("error - ", err)
		}
		fmt.Println("k:", k, "v:", v)
		i = i + 1
		// time.Sleep(100 * time.Millisecond)
	}
}

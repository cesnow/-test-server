package main

import (
	"context"
	"fmt"
	kafka "kiyudesign.com/cesnow/light-server/pkg/mq"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	job := kafka.MustKafkaConsumer(&kafka.KafkaConsumerConf{
		Topics:  []string{"chat-test-topic"},
		Brokers: []string{"127.0.0.1:9092"},
		Group:   "chat-test-group-job",
	})

	job.RegisterHandlers("chat-test-topic",
		func(ctx context.Context, method, key string, value []byte) {
			fmt.Println("key: ", key, ", value: ", string(value))
		})

	defer job.Stop()
	go job.Start()
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		fmt.Println("get a signal ", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// job.Close()
			fmt.Println("exit...")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

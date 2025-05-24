package kafka

import (
	"context"
	"fmt"
	"kiyudesign.com/cesnow/light-server/pkg/error2"

	"github.com/IBM/sarama"
)

func Check(ctx context.Context, conf *KafkaConsumerConf, topics []string) error {
	kfk, err := BuildConsumerGroupConfig(conf, sarama.OffsetNewest, false)
	if err != nil {
		return err
	}
	cli, err := sarama.NewClient(conf.Brokers, kfk)
	if err != nil {
		return error2.Wrapf(err, "NewClient failed - {config: %v}", fmt.Sprintf("%+v", conf))
	}
	defer cli.Close()

	existingTopics, err := cli.Topics()
	if err != nil {
		return error2.Wrap(err, "Failed to list topics")
	}

	existingTopicsMap := make(map[string]bool)
	for _, t := range existingTopics {
		existingTopicsMap[t] = true
	}

	for _, topic := range topics {
		if !existingTopicsMap[topic] {
			return error2.Wrap(fmt.Errorf("topic not exist - {topic: %v}", topic), "")
		}
	}
	return nil
}

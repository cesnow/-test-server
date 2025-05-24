package kafka

import (
	"context"
	"errors"
	"kiyudesign.com/cesnow/light-server/pkg/error2"

	"github.com/IBM/sarama"
)

var errEmptyMsg = errors.New("binary msg is empty")

type Producer struct {
	producer sarama.SyncProducer
	c        *KafkaProducerConf
}

func MustKafkaProducer(c *KafkaProducerConf) *Producer {
	conf, err := BuildProducerConfig(*c)
	if err != nil {
		panic(err)
	}

	producer, err := NewProducer(conf, c.Brokers)
	if err != nil {
		panic(err)
	}

	return &Producer{
		producer: producer,
		c:        c}
}

// SendMessage
// Input send msg to kafka
// NOTE: If producer has beed created failed, the message will lose.
func (p *Producer) SendMessage(ctx context.Context, key string, value []byte) (partition int32, offset int64, err error) {
	ctx, span := startProducerSpan(ctx, "SendMessage")
	defer func() {
		endProducerSpan(span, err)
	}()

	if len(value) == 0 {
		err = error2.Wrapf(errors.New("len(value) == 0 "), "")
		return
	}

	// Prepare Kafka message
	kMsg := &sarama.ProducerMessage{
		Topic: p.Topic(),
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// Validate message key and value
	if kMsg.Key.Length() == 0 || kMsg.Value.Length() == 0 {
		err = error2.Wrap(errEmptyMsg, "")
		return
	}

	// Attach context metadata as headers
	kMsg.Headers = extractTraceHeaders(ctx)

	// Send the message
	partition, offset, err = p.producer.SendMessage(kMsg)
	if err != nil {
		err = error2.Wrapf(err, "p.producer.SendMessage error")
	}

	return
}

func (p *Producer) SendMessageV2(ctx context.Context, method, key string, value []byte) (partition int32, offset int64, err error) {
	ctx, span := startProducerSpan(ctx, "SendMessage")
	defer func() {
		endProducerSpan(span, err)
	}()

	if len(value) == 0 {
		err = error2.Wrapf(errors.New("len(value) == 0 "), "")
		return
	}

	// Prepare Kafka message
	kMsg := &sarama.ProducerMessage{
		Topic: p.Topic(),
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// Validate message key and value
	if kMsg.Key.Length() == 0 || kMsg.Value.Length() == 0 {
		err = error2.Wrap(errEmptyMsg, "")
		return
	}

	// Attach context metadata as headers
	kMsg.Headers = append([]sarama.RecordHeader{{Key: []byte("method"), Value: []byte(method)}}, extractTraceHeaders(ctx)...)

	// Send the message
	partition, offset, err = p.producer.SendMessage(kMsg)
	if err != nil {
		err = error2.Wrapf(err, "p.producer.SendMessage error")
	}

	return
}

func (p *Producer) Close() (err error) {
	if p.producer != nil {
		return p.producer.Close()
	}
	return
}

func (p *Producer) Topic() string {
	return p.c.Topic
}

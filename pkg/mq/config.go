package kafka

import (
	"bytes"
	"kiyudesign.com/cesnow/light-server/pkg/error2"
	"strings"

	"github.com/IBM/sarama"
)

type TLSConfig struct {
	EnableTLS          bool   `json:",optional"`
	CACrt              string `json:",optional"`
	ClientCrt          string `json:",optional"`
	ClientKey          string `json:",optional"`
	ClientKeyPwd       string `json:",optional"`
	InsecureSkipVerify bool   `json:",optional"`
}

// KafkaProducerConf kafka producer settings.
type KafkaProducerConf struct {
	Username     string    `json:",optional"`
	Password     string    `json:",optional"`
	ProducerAck  string    `json:",optional"`
	CompressType string    `json:",optional"`
	TLS          TLSConfig `json:",optional"`
	Brokers      []string
	Topic        string
}

// KafkaConsumerConf kafka client settings.
type KafkaConsumerConf struct {
	Username     string    `json:",optional"`
	Password     string    `json:",optional"`
	ProducerAck  string    `json:",optional"`
	CompressType string    `json:",optional"`
	TLS          TLSConfig `json:",optional"`
	Brokers      []string
	Topics       []string
	Group        string
}

// KafkaBatchConsumerConf kafka client settings.
type KafkaBatchConsumerConf struct {
	KafkaConsumerConf
	Duration   int `json:",default=100"`
	ChannelNum int `json:"default=50"`
}

type KafkaShardingConsumerConf struct {
	KafkaConsumerConf
	Concurrency int    `json:",default=2048"`
	QueueBuffer int    `json:",default=128"`
	ClientId    string `json:",default=sarama"`
}

func BuildConsumerGroupConfig(conf *KafkaConsumerConf, initial int64, autoCommitEnable bool) (*sarama.Config, error) {
	kfk := sarama.NewConfig()
	kfk.Version = sarama.V2_0_0_0
	kfk.Consumer.Offsets.Initial = initial
	kfk.Consumer.Offsets.AutoCommit.Enable = autoCommitEnable
	kfk.Consumer.Return.Errors = false
	if conf.Username != "" || conf.Password != "" {
		kfk.Net.SASL.Enable = true
		kfk.Net.SASL.User = conf.Username
		kfk.Net.SASL.Password = conf.Password
	}
	if conf.TLS.EnableTLS {
		tls, err := newTLSConfig(conf.TLS.ClientCrt, conf.TLS.ClientKey, conf.TLS.CACrt, []byte(conf.TLS.ClientKeyPwd), conf.TLS.InsecureSkipVerify)
		if err != nil {
			return nil, err
		}
		kfk.Net.TLS.Config = tls
		kfk.Net.TLS.Enable = true
	}
	return kfk, nil
}

func NewConsumerGroup(conf *sarama.Config, addr []string, groupID string) (sarama.ConsumerGroup, error) {
	cg, err := sarama.NewConsumerGroup(addr, groupID, conf)
	if err != nil {
		return nil, error2.Wrapf(err, "NewConsumerGroup failed: {addr: %v, groupID: %v, conf: %v", addr, groupID, *conf)
	}
	return cg, nil
}

func BuildProducerConfig(conf KafkaProducerConf) (*sarama.Config, error) {
	kfk := sarama.NewConfig()
	kfk.Producer.Return.Successes = true
	kfk.Producer.Return.Errors = true
	kfk.Producer.Partitioner = sarama.NewHashPartitioner
	if conf.Username != "" || conf.Password != "" {
		kfk.Net.SASL.Enable = true
		kfk.Net.SASL.User = conf.Username
		kfk.Net.SASL.Password = conf.Password
	}
	switch strings.ToLower(conf.ProducerAck) {
	case "no_response":
		kfk.Producer.RequiredAcks = sarama.NoResponse
	case "wait_for_local":
		kfk.Producer.RequiredAcks = sarama.WaitForLocal
	case "wait_for_all":
		kfk.Producer.RequiredAcks = sarama.WaitForAll
	default:
		kfk.Producer.RequiredAcks = sarama.WaitForAll
	}
	if conf.CompressType == "" {
		kfk.Producer.Compression = sarama.CompressionNone
	} else {
		if err := kfk.Producer.Compression.UnmarshalText(bytes.ToLower([]byte(conf.CompressType))); err != nil {
			return nil, error2.Wrapf(err, "UnmarshalText failed: {compressType: %v}", conf.CompressType)
		}
	}
	if conf.TLS.EnableTLS {
		tls, err := newTLSConfig(conf.TLS.ClientCrt, conf.TLS.ClientKey, conf.TLS.CACrt, []byte(conf.TLS.ClientKeyPwd), conf.TLS.InsecureSkipVerify)
		if err != nil {
			return nil, err
		}
		kfk.Net.TLS.Config = tls
		kfk.Net.TLS.Enable = true
	}
	return kfk, nil
}

func NewProducer(conf *sarama.Config, addr []string) (sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(addr, conf)
	if err != nil {
		return nil, error2.Wrapf(err, "NewSyncProducer failed - {addr: %v, conf: %v}", addr, *conf)
	}
	return producer, nil
}

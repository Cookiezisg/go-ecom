package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/oklog/ulid/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

// Kafka配置
type Config struct {
	Brokers       []string `json:"required"`
	ProducerAsync bool     `json:"default=true"`
	Version       string   `json:"default=2.8.0"`
	ConsumerGroup string   `json:"optional"`
}

// Message 消息结构
type Message struct {
	Version   string                 `json:"version"`
	MessageID string                 `json:"message_id"`
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}

func NewMessage(eventType string, data map[string]interface{}) *Message {
	return &Message{
		Version:   "1.0",
		MessageID: generateMessageID(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Data:      data,
	}
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

type Producer struct {
	producer sarama.AsyncProducer
	config   *Config
}

func NewProducer(cfg *Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy

	kafkaVersion, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		kafkaVersion = sarama.V2_8_0_0
	}
	config.Version = kafkaVersion

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka生产者失败: %w", err)
	}

	p := &Producer{
		producer: producer,
		config:   cfg,
	}

	go p.handleErrors()
	go p.handleSuccesses()

	return p, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("消息序列化失败: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.MessageID), // 使用消息ID作为key，保证有序
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(message.EventType),
			},
			{
				Key:   []byte("message_id"),
				Value: []byte(message.MessageID),
			},
		},
		Timestamp: time.Now(),
	}

	select {
	case p.producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("消息发送超时: %w", ctx.Err())
	}
}

// handleErrors 处理错误
func (p *Producer) handleErrors() {
	for err := range p.producer.Errors() {
		logx.Errorf("Kafka生产者错误: topic=%s, error=%v", err.Msg.Topic, err.Err)
	}
}

// handleSuccesses 处理成功消息
func (p *Producer) handleSuccesses() {
	for msg := range p.producer.Successes() {
		logx.Infof("Kafka消息发送成功: topic=%s, partition=%d, offset=%d",
			msg.Topic, msg.Partition, msg.Offset)
	}
}

// 定义一个池或者全局变量来保证并发安全和性能
var (
	entropy   *ulid.MonotonicEntropy
	entropyMu sync.Mutex
)

func init() {
	// 初始化熵源（随机数种子）
	t := time.Now()
	entropy = ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
}

// generateMessageID 现在产出的是优雅且专业的 ULID
func generateMessageID() string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	// 使用当前时间生成 ULID
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return id.String()
}

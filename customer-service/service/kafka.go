package service

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

// KafkaProducer is a global Kafka producer instance
var KafkaProducer sarama.SyncProducer

// InitializeKafkaProducer initializes the Kafka producer with the given broker addresses
func InitializeKafkaProducer(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Ensure all brokers acknowledge
	config.Producer.Retry.Max = 5                    // Retry up to 5 times
	config.Producer.Return.Successes = true          // Return success metadata

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}

	KafkaProducer = producer
	return nil
}

// SendMessage sends a JSON-encoded message to the specified Kafka topic
func SendMessage(topic string, message interface{}) error {
	// Marshal the message into JSON
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Create a Kafka producer message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msgBytes),
	}

	// Send the message
	_, _, err = KafkaProducer.SendMessage(msg)
	return err
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/IBM/sarama"
	"gopkg.in/gomail.v2"
)

type CustomerEvent struct {
	UserID   string `json:"UserID"`
	Name     string `json:"Name"`
	Phone    string `json:"Phone"`
	Address  string `json:"Address"`
	Address2 string `json:"Address2"`
	City     string `json:"City"`
	State    string `json:"State"`
	Zipcode  string `json:"Zipcode"`
}

func main() {
	brokers := []string{os.Getenv("KAFKA_BROKER")} // Use environment variable for Kafka broker
	topic := os.Getenv("KAFKA_TOPIC")              // Use environment variable for Kafka topic

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	log.Printf("Listening for messages on topic: %s", topic)

	for message := range partitionConsumer.Messages() {
		var event CustomerEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		sendEmail(event)
	}
}

func sendEmail(event CustomerEvent) {
	email := os.Getenv("SMTP_EMAIL")       // Use environment variable for email
	password := os.Getenv("SMTP_PASSWORD") // Use environment variable for email password

	dialer := gomail.NewDialer("smtp.gmail.com", 587, email, password)

	msg := gomail.NewMessage()
	msg.SetHeader("From", email)
	msg.SetHeader("To", event.UserID) // Send email to the customer's email address
	msg.SetHeader("Subject", "Activate your book store account")
	msg.SetBody("text/plain", fmt.Sprintf(
		"Dear %s,\n\nWelcome to the Book store created by <your-andrew-ID>.\n\nExceptionally this time we won’t ask you to click a link to activate your account.",
		event.Name,
	))

	if err := dialer.DialAndSend(msg); err != nil {
		log.Printf("Failed to send email: %v", err)
	} else {
		log.Printf("Email sent to %s", event.Name)
	}
}

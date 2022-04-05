package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Serialize(url1 string, url2 string) []byte {
	bt, _ := json.Marshal([]string{url1, url2})
	return bt
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"workers", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	resultsQueue, err := ch.QueueDeclare(
		"results", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		resultsQueue.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Print("Failed to register a consumer")
	}

	forever := make(chan bool)

	go func() {
		for v := range msgs {
			var data []string

			json.Unmarshal(v.Body, &data)

			log.Printf("FROM: %s	TO:		%s", data[0], data[1])

			for i, urls := range data {
				if i > 1 {
					log.Printf("%s --->", urls)
				}
			}
		}
	}()

	url_pairs := [][]string{{"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Help:Introduction"},
		{"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Carleton_College"},
		{"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Search_for_extraterrestrial_intelligence"},
		{"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Archaeology,_Anthropology,_and_Interstellar_Communication"}}
	for _, v := range url_pairs {
		body := Serialize(v[0], v[1])
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s\n", body)
	}
	<-forever
}

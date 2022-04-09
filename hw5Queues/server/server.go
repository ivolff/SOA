package main

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var submited_urls = map[string][]string{}

// "1": {"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Help:Introduction"},
// "2": {"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Carleton_College"},
// "3": {"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Search_for_extraterrestrial_intelligence"},
// "4": {"https://en.wikipedia.org/wiki/Talk:Main_Page", "https://en.wikipedia.org/wiki/Archaeology,_Anthropology,_and_Interstellar_Communication"}}

var results map[string][]string

var results_mutex sync.Mutex

var nextId int64

var submited_urls_mutex sync.Mutex

func submitTask(request [2]string) string {
	submited_urls_mutex.Lock()
	id := atomic.AddInt64(&nextId, 1)
	submited_urls[string(id)] = []string{request[0], request[1]}
	submited_urls_mutex.Unlock()
	return string(id)
}

func serveRPC(conn *amqp.Connection) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a RPC channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a RPC queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set RPC QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a RPC consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var request [2]string
			json.Unmarshal(d.Body, &request)

			failOnError(err, "Failed to convert body to integer")

			id := submitTask(request)

			for {
				if results[id] != nil {
					results_mutex.Lock()
					responce := results[id]
					delete(results, id)
					results_mutex.Unlock()

					body, _ := json.Marshal(responce)

					err = ch.Publish(
						"",        // exchange
						d.ReplyTo, // routing key
						false,     // mandatory
						false,     // immediate
						amqp.Publishing{
							ContentType:   "text/plain",
							CorrelationId: d.CorrelationId,
							Body:          []byte(body),
						})
					failOnError(err, "Failed to publish a message")

					d.Ack(false)
					break
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}

func Serialize(id int64, url1 string, url2 string) []byte {
	bt, _ := json.Marshal([]string{string(id), url1, url2})
	return bt
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	results = make(map[string][]string)

	go func() {
		for v := range msgs {
			var data []string

			log.Println(string(v.Body))

			json.Unmarshal(v.Body, &data)

			log.Printf("FROM: %s	TO:		%s", data[1], data[2])

			for i, urls := range data {
				if i > 3 {
					log.Printf("%s --->", urls)
				}
			}

			results_mutex.Lock()
			results[data[0]] = data[3:]
			results_mutex.Unlock()

		}
	}()

	go func() {
		for {
			local_submited_urls := make(map[string][]string)

			submited_urls_mutex.Lock()
			for key, value := range submited_urls {
				local_submited_urls[key] = value
			}
			submited_urls = make(map[string][]string)

			submited_urls_mutex.Unlock()

			for _, v := range local_submited_urls {
				body := Serialize(nextId, v[0], v[1])
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
			time.Sleep(2 * time.Second)
			log.Printf("waiting for tasks")
		}
	}()

	go serveRPC(conn)

	<-forever
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sync/atomic"

	"github.com/streadway/amqp"
	"golang.org/x/net/html"
)

var ignoredPatterns = map[string]bool{
	"#": true,
	"/": true,
}

func parceLinks(resp *http.Response, currentURL string) []string {
	result := make([]string, 0)

	u, err := url.Parse(currentURL)

	if err != nil {
		fmt.Errorf("Invalid URL")
		return result
	}

	hostUrl := "http://" + u.Host

	tokenizer := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokenizer.Next()

		switch {
		case tokenType == html.ErrorToken:
			return result
		case tokenType == html.StartTagToken:
			token := tokenizer.Token()

			if token.Data == "a" {
				for _, a := range token.Attr {
					if a.Key == "href" && !ignoredPatterns[a.Val] {
						link := a.Val
						if len(a.Val) > 3 && a.Val[:4] != "http" {
							link = hostUrl + link
						}

						u1, erru := url.Parse(link)

						if erru == nil && (u1.Host+u1.Path) != "" {
							//fmt.Println(u1.Scheme, u1.Opaque, u1.User, u1.Host, u1.Path, u1.RawPath, u1.ForceQuery, u1.RawQuery, u1.Fragment, u1.RawFragment)
							result = append(result, u1.Scheme+"://"+u1.Host+u1.Path)
							break
						}
					}
				}
			}
		}
	}
}

func isUrlsEqual(one string, two string) bool {
	u1, err1 := url.Parse(one)
	u2, err2 := url.Parse(two)

	if err1 != nil {
		fmt.Println("err1", one)
		return false
	}

	if err2 != nil {
		fmt.Println("err2", two)
		return false
	}

	// if u1.Host == u2.Host && u1.Path == u2.Path {
	// 	fmt.Println(u1.Host, u2.Host, u1.Path, u2.Path, u1.Host == u2.Host, u1.Path == u2.Path)
	// }
	//fmt.Println(u1.Host, u2.Host, u1.Path, u2.Path, u1.Host == u2.Host, u1.Path == u2.Path)
	return ((u1.Host == u2.Host) && (u1.Path == u2.Path))
}

func CopyMap(Map *map[string]int) map[string]int {
	Res := make(map[string]int, 0)
	for k, v := range *Map {
		Res[k] = v
	}
	return Res
}

var MaxDepth = 10 //максимальная глубина на которую ищем
var ResultTrace = make(map[string]int)
var m atomic.Value //0 - когда не нашли ни одного пути 1 иначе
var counter int64  //считаем количество горутин

func crawler(Trace *map[string]int, Url string, Target string, depth int) {

	atomic.AddInt64(&counter, 1)

	CurTrace := CopyMap(Trace)

	CurTrace[Url] = depth
	URL := strings.TrimSpace(Url)
	Resp, err := http.Get(URL)

	if err != nil {
		fmt.Println("Request error:", err)
		onCrawlerEnd()
		return
	}

	urls := parceLinks(Resp, URL)
	isFounded := false
	for _, ur := range urls {
		if isUrlsEqual(ur, Target) {
			if m.CompareAndSwap(0, 1) {
				CurTrace[ur] = depth + 1
				//fmt.Println("FIND!!!", CurTrace)
				ResultTrace = CurTrace
			}
			onCrawlerEnd()
			return
		}
		if m.Load() == 1 {
			break
		}
		if !(CurTrace[ur] > 0) && (depth < MaxDepth) && !isFounded {
			go crawler(&CurTrace, ur, Target, depth+1)
			if isFounded {
				onCrawlerEnd()
				return
			}
		}
	}
	onCrawlerEnd()
	return
}

func onCrawlerEnd() {
	atomic.AddInt64(&counter, -1)
}

func main() {
	//pid := os.Args[1]
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
	if err != nil {
		fmt.Print("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Print("Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"workers", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		fmt.Print("Failed to declare a queue")
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		fmt.Print("Failed to register a consumer")
	}

	results, err := ch.QueueDeclare(
		"results", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		fmt.Print("Failed to declare a queue")
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			ResultTrace = make(map[string]int)
			atomic.StoreInt64(&counter, 0)
			var data [2]string
			json.Unmarshal(d.Body, &data)
			//log.Printf("Received a message: %s,   %s", data[0], data[1])

			firstURL := strings.TrimSpace(data[0])
			secondURL := strings.TrimSpace(data[1])

			//respFirst, _ := http.Get(firstURL)
			Trace := make(map[string]int, 0)

			m.Store(0)

			crawler(&Trace, firstURL, secondURL, 0)

			time.Sleep(360) //Даем наспавниться горутинам
			for len(ResultTrace) == 0 && atomic.LoadInt64(&counter) != 0 {
				time.Sleep(360)
				//fmt.Println(len(ResultTrace), atomic.LoadInt64(&counter))
			}

			SortedTrace := make([]string, len(ResultTrace)+2)

			for k, v := range ResultTrace {
				SortedTrace[v+2] = k
			}
			//log.Println("TRACE")
			//log.Print(SortedTrace)
			SortedTrace[0] = firstURL
			SortedTrace[1] = secondURL
			body, _ := json.Marshal(SortedTrace)

			err = ch.Publish(
				"",           // exchange
				results.Name, // routing key
				false,        // mandatory
				false,        // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			if err != nil {
				fmt.Print("Failed to publish")
			}
			//log.Printf(" [x] Sent %s\n", body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

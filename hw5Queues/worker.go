package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"sync/atomic"

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

	if u1.Host == u2.Host && u1.Path == u2.Path {
		fmt.Println(u1.Host, u2.Host, u1.Path, u2.Path, u1.Host == u2.Host, u1.Path == u2.Path)
	}
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

var MaxDepth = 3
var ResultTrace = make(map[string]int)
var m atomic.Value

func crawler(Trace *map[string]int, Url string, Target string, depth int) {

	CurTrace := CopyMap(Trace)

	CurTrace[Url] = depth
	URL := strings.TrimSpace(Url)
	Resp, err := http.Get(URL)

	if err != nil {
		return
	}

	urls := parceLinks(Resp, URL)
	isFounded := false
	for _, ur := range urls {
		if isUrlsEqual(ur, Target) {
			if m.CompareAndSwap(0, 1) {
				CurTrace[ur] = depth + 1
				fmt.Println("FIND!!!", CurTrace)
				ResultTrace = CurTrace
			}
			return
		}
		if m.Load() == 1 {
			break
		}
		if !(CurTrace[ur] > 0) && (depth < MaxDepth) && !isFounded {
			go crawler(&CurTrace, ur, Target, depth+1)
			if isFounded {
				return
			}
		}
	}
	return
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	firstURL, _ := reader.ReadString('\n')
	secondURL, _ := reader.ReadString('\n')

	firstURL = strings.TrimSpace(firstURL)
	secondURL = strings.TrimSpace(secondURL)

	//respFirst, _ := http.Get(firstURL)
	Trace := make(map[string]int, 0)

	m.Store(0)

	crawler(&Trace, firstURL, secondURL, 0)

	for len(ResultTrace) == 0 {
		fmt.Println(len(ResultTrace))
	}

	SortedTrace := make([]string, len(ResultTrace))

	for k, v := range ResultTrace {
		SortedTrace[v] = k
	}

	fmt.Print(SortedTrace)
}

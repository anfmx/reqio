package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	method    = flag.String("m", "GET", "Current method\n")
	showBody  = flag.Bool("b", false, "show response body\n")
	limit     = flag.Int("limit", 0, "sets limit to body\n")
	rate      = flag.Int("rate", 0, "sends request per N second\n")
	writeFile = flag.Bool("f", false, "")
	Client    = &http.Client{}
)

func main() {
	flag.Parse()
	Execute()
}

func Execute() {
	if *rate == 0 {
		doRequest()
		return
	}
	doRequestLoop()
}

func doRequestLoop() {
	ticker := time.NewTicker(time.Duration(*rate) * time.Second)
	defer ticker.Stop()

	// TODO: Replace with `select` and context cancellation
	//
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 	}
	// }
	for range ticker.C {
		go doRequest()
	}
}

func urlChanGenerator() chan string {
	urlsChan := make(chan string)
	urls := flag.Args()

	go func() {
		for i := range urls {
			urlsChan <- urls[i]
		}
		close(urlsChan)
	}()

	return urlsChan
}

func doRequest() {
	urlsChan := urlChanGenerator()
	file, _ := os.OpenFile("requests.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0744)
	defer file.Close()

	for {
		select {
		case url, ok := <-urlsChan:
			if !ok {
				return
			}

			req, err := http.NewRequest(*method, url, nil)
			if err != nil {
				fmt.Println(err)
			}

			resp, err := Client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()

			fmt.Println(resp.Status)
			if *writeFile {
				fmt.Fprintln(file, resp.Status)
			}

			if *showBody {
				body := []interface{}{}
				json.NewDecoder(resp.Body).Decode(&body)
				displayBody(body, file)
			}
		}
	}
}

func displayBody(body []interface{}, file *os.File) {
	if len(body) == 0 {
		fmt.Println("Non-existent url or no response body")
		return
	}

	if *limit != 0 && *limit < len(body) {
		body = body[:*limit]
	}

	prettyBody, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(prettyBody))
	if *writeFile {
		fmt.Fprintln(file, string(prettyBody))
	}
}

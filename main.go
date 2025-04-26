package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	fileWriteChan := make(chan string)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		FileWriter(fileWriteChan)
	}()

	urlsChan := urlChanGenerator(ctx)
	if *rate == 0 {
		doRequest(urlsChan, fileWriteChan)
	} else {
		doRequestLoop(urlsChan, fileWriteChan, ctx)
	}
	close(fileWriteChan)
	wg.Wait()
}

// FIXME: writes once then context deny
func doRequestLoop(urlsChan chan string, fileWriteChan chan string, ctx context.Context) {
	ticker := time.NewTicker(time.Duration(*rate) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go doRequest(urlsChan, fileWriteChan)
		case <-ctx.Done():
			return
		}
	}
}

func urlChanGenerator(ctx context.Context) chan string {
	urlsChan := make(chan string)
	urls := flag.Args()

	go func() {
		for _, url := range urls {
			urlsChan <- url
		}
		close(urlsChan)
	}()

	return urlsChan
}

func FileWriter(fileWriteChan chan string) {
	if !*writeFile {
		return
	}

	file, err := os.OpenFile("requests.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0744)
	if err != nil {
		fmt.Println("Cannot open a file to write logs", err)
		*writeFile = false
		return
	}
	defer file.Close()

	for data := range fileWriteChan {
		file.WriteString(data)
	}
}

func doRequest(urlsChan chan string, fileWriteChan chan string) {
	var wg sync.WaitGroup

	for url := range urlsChan {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			req, err := http.NewRequest(*method, url, nil)
			if err != nil {
				fmt.Println(err)
			}
			req = req.WithContext(ctx)

			resp, err := Client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()

			var output strings.Builder
			output.WriteString(fmt.Sprintf("%s | %s\n", req.URL, resp.Status))

			if *showBody {
				body := []interface{}{}
				json.NewDecoder(resp.Body).Decode(&body)

				if len(body) != 0 {
					if *limit != 0 && *limit < len(body) {
						body = body[:*limit]
					}
					prettyBody, err := json.MarshalIndent(body, "", " ")
					if err != nil {
						fmt.Println(err)
					}
					output.WriteString(string(prettyBody) + "\n")
				}
			}
			fmt.Println(output.String())
			if *writeFile {
				fileWriteChan <- output.String()
			}
		}(url)
	}
	wg.Wait()
}

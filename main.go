package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	rateLimit = flag.Int("rate-limit", 10, "Use with'--rate' flag to limit overall time of making requests")
	bodyData  = flag.String("d", "", "Needs to send request with json body")

	Client = &http.Client{}
)

type jsonData struct {
	Url        string      `json:"url"`
	StatusCode string      `json:"status_code"`
	Body       interface{} `json:"body"`
}

func NewJsonData(url, code string, body interface{}) jsonData {
	return jsonData{
		Url:        url,
		StatusCode: code,
		Body:       body,
	}
}

func main() {
	flag.Parse()
	Execute()
}

func Execute() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*rateLimit))
	defer cancel()

	fileWriteChan := make(chan string)

	wg := &sync.WaitGroup{}

	// NOTE: To avoid unnecessary handoff there is an additional check
	if *writeFile {
		wg.Add(1)
		go FileWriter(fileWriteChan, wg, ctx)
	}

	urlsChan := urlChanGenerator(ctx)
	if *rate == 0 {
		doRequest(urlsChan, fileWriteChan)
	} else {
		doRequestLoop(fileWriteChan, ctx)
	}

	wg.Wait()
}

func doRequestLoop(fileWriteChan chan string, ctx context.Context) {
	ticker := time.NewTicker(time.Duration(*rate) * time.Second)
	defer ticker.Stop()

	urls := flag.Args()
	if len(urls) == 0 {
		return
	}

	for {
		select {
		case <-ticker.C:

			innerFileWriteChan := make(chan string)

			go func() {
				for data := range innerFileWriteChan {
					select {
					case fileWriteChan <- data:
					case <-ctx.Done():
						return
					}
				}
			}()

			urlsChan := urlChanGenerator(ctx)
			doRequest(urlsChan, innerFileWriteChan)

		case <-ctx.Done():
			close(fileWriteChan)
			return
		}
	}
}

func urlChanGenerator(ctx context.Context) chan string {
	urlsChan := make(chan string)
	urls := flag.Args()

	go func() {
		for _, url := range urls {
			select {
			case urlsChan <- url:
			case <-ctx.Done():
				return
			}
		}
		close(urlsChan)
	}()
	return urlsChan
}

func FileWriter(fileWriteChan chan string, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	file, err := os.OpenFile("requests.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0744)
	if err != nil {
		fmt.Println("Cannot open a file to write logs", err)
		*writeFile = false
		return
	}
	defer file.Close()

	for {
		select {
		case data, ok := <-fileWriteChan:
			if !ok {
				return
			}
			file.WriteString(data)
		case <-ctx.Done():
			return
		}
	}
}

func doRequest(urlsChan chan string, fileWriteChan chan string) {
	wg := &sync.WaitGroup{}

	for url := range urlsChan {
		wg.Add(1)
		go handleRequest(url, fileWriteChan, wg)
	}

	wg.Wait()
	close(fileWriteChan)
}

func handleRequest(url string, fileWriteChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := newRequest(url, ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	output, err := processResponse(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(output)

	if *writeFile {
		select {
		case fileWriteChan <- output:
		case <-ctx.Done():
			return
		}
	}
}

func processResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	var body interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	jsonResult := NewJsonData(resp.Request.URL.String(), resp.Status, body)
	prettyJson, err := json.MarshalIndent(jsonResult, "", " ")
	if err != nil {
		return "", err
	}

	return string(prettyJson) + "\n", nil
}

func newRequest(url string, ctx context.Context) (*http.Request, error) {
	var bodyReader io.Reader
	if *bodyData != "" {
		bodyReader = strings.NewReader(*bodyData)
	}
	req, err := http.NewRequest(*method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if *bodyData != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req.WithContext(ctx), nil
}

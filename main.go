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
		go FileWriter(fileWriteChan, wg)
	}

	urlsChan := urlChanGenerator(ctx)
	if *rate == 0 {
		doRequest(urlsChan, fileWriteChan)
	} else {
		doRequestLoop(fileWriteChan, ctx)
	}

	wg.Wait()
	close(fileWriteChan)
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
			urlChan := urlChanGenerator(ctx)
			doRequest(urlChan, fileWriteChan)
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

func FileWriter(fileWriteChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()

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
	wg := &sync.WaitGroup{}

	for url := range urlsChan {
		wg.Add(1)
		go handleRequest(url, fileWriteChan, wg)
	}
	wg.Wait()
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

	var output strings.Builder
	output.WriteString(fmt.Sprintf("%s | %s\n", resp.Request.URL, resp.Status))

	if *showBody {
		err := processBody(&output, resp)
		if err != nil {
			return "", err
		}
	}
	return output.String(), nil
}

func processBody(output *strings.Builder, resp *http.Response) error {
	body := []interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	if len(body) != 0 {
		if *limit != 0 && *limit < len(body) {
			body = body[:*limit]
		}
		prettyBody, err := json.MarshalIndent(body, "", " ")
		if err != nil {
			return err
		}
		output.WriteString(string(prettyBody) + "\n\n")
	}
	return nil
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

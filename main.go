package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

var (
	method   = flag.String("m", "GET", "Current method\n")
	showBody = flag.Bool("b", false, "show response body\n")
	limit    = flag.Int("limit", 0, "sets limit to body\n")
	rate     = flag.Int("rate", 0, "sends request per N second\n")

	Client = &http.Client{}
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
			fmt.Println(resp.Status)
			if *showBody {
				displayBody(resp, []interface{}{})
			}
		}
	}

}

func displayBody(resp *http.Response, res []interface{}) {
	json.NewDecoder(resp.Body).Decode(&res)

	if len(res) == 0 {
		fmt.Println("Non-existent url or no response body")
		return
	}

	if *limit != 0 && *limit < len(res) {
		prettyBody, err := json.MarshalIndent(res[:*limit], "", " ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(prettyBody))
		return
	}

	prettyBody, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(prettyBody))
}

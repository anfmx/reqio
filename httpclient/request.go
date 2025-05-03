package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reqio/utils"
	"strings"
	"sync"
	"time"
)

type DefaultRequester struct {
	Client *http.Client
	Config *utils.Config
	Urls   []string
}

func (dr DefaultRequester) DoRequest(urlsChan chan string, fileWriteChan chan string) {
	wg := &sync.WaitGroup{}

	for url := range urlsChan {
		wg.Add(1)
		go dr.handleRequest(url, fileWriteChan, wg)
	}

	wg.Wait()
	close(fileWriteChan)
}

func (dr DefaultRequester) DoRequestLoop(fileWriteChan chan string, ctx context.Context) {
	ticker := time.NewTicker(time.Duration(dr.Config.Rate) * time.Second)
	defer ticker.Stop()

	if len(dr.Urls) == 0 {
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

			urlsChan := dr.UrlChanGenerator(ctx)
			dr.DoRequest(urlsChan, innerFileWriteChan)

		case <-ctx.Done():
			close(fileWriteChan)
			return
		}
	}
}

func (dr DefaultRequester) newRequest(url string, ctx context.Context) (*http.Request, error) {
	var bodyReader io.Reader
	if dr.Config.BodyData != "" {
		bodyReader = strings.NewReader(dr.Config.BodyData)
	}
	req, err := http.NewRequest(dr.Config.Method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if dr.Config.BodyData != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req.WithContext(ctx), nil
}

func (dr DefaultRequester) handleRequest(url string, fileWriteChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := dr.newRequest(url, ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := dr.Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	output, err := utils.ProcessResponse(resp, dr.Config.ShowBody, dr.Config.Limit)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(output)

	if dr.Config.WriteFile {
		select {
		case fileWriteChan <- output:
		case <-ctx.Done():
			return
		}
	}
}

func (dr DefaultRequester) UrlChanGenerator(ctx context.Context) chan string {
	urlsChan := make(chan string)
	go func() {
		for _, url := range dr.Urls {
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

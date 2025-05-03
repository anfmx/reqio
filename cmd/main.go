package main

import (
	"context"
	"flag"
	"net/http"
	"reqio/httpclient"
	"reqio/logger"
	"reqio/utils"
	"sync"
	"time"
)

func main() {
	cfg := utils.ParseFlags()
	urls := flag.Args()

	client := httpclient.DefaultRequester{
		Client: &http.Client{},
		Config: cfg,
		Urls:   urls,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.TimeLimit))
	defer cancel()

	fileWriteChan := make(chan string)

	wg := &sync.WaitGroup{}
	if cfg.WriteFile {
		wg.Add(1)
		go logger.FileWriter(fileWriteChan, wg, ctx)
	}

	urlsChan := client.UrlChanGenerator(ctx)
	if cfg.Rate == 0 {
		client.DoRequest(urlsChan, fileWriteChan)

	} else {
		client.DoRequestLoop(fileWriteChan, ctx)
	}

	wg.Wait()
}

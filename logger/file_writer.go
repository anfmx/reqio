package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
)

func FileWriter(fileWriteChan chan string, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	file, err := os.OpenFile("requests.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0744)
	if err != nil {
		fmt.Println("Cannot open a file to write logs", err)
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

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// this will end up as a CLI arg but hardcoded for now
const RequestPerSecond = 0.2

func make_request() error {
	resp, err := http.Get("http://127.0.0.1:8080")
	if err != nil {
		log.Fatal("failed make request", "error", err)
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("failed read body", "error", err)
	}
	fmt.Println("response:", string(body))
	return nil
}

func getDuration() time.Duration {
	return time.Duration(float64(time.Second) / RequestPerSecond)
}

func main() {

	for {
		tm := time.Now().Format(time.RFC1123)
		fmt.Println("The time is: " + tm)
		err := make_request()
		if err != nil {
			log.Fatal("request failed", "error", err)
		}
		time.Sleep(getDuration())
	}
}

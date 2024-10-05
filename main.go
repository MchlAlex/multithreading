package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const cep = "01153000"

func main() {
	// Set timeout of 1 second
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan struct {
		Response []byte
		API      string
		Error    error
	}, 2)

	// Concurrent API requests
	go fetchFromAPI(ctx, fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep), "BrasilAPI", ch)
	go fetchFromAPI(ctx, fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep), "ViaCEP", ch)

	// Process the fastest response or timeout
	select {
	case result := <-ch:
		if result.Error != nil {
			log.Printf("Error: %v", result.Error)
		} else {
			// Print the raw JSON response and which API responded
			fmt.Printf("Response from %s:\n%s\n", result.API, string(result.Response))
			cancel()
		}
	case <-ctx.Done():
		log.Println("Timeout: No response within 1 second.")
	}
}

func fetchFromAPI(ctx context.Context, url, apiName string, ch chan<- struct {
	Response []byte
	API      string
	Error    error
},
) {
	// Create a new request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- struct {
			Response []byte
			API      string
			Error    error
		}{Error: err}
		return
	}

	// Make the API request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- struct {
			Response []byte
			API      string
			Error    error
		}{Error: err}
		return
	}
	defer res.Body.Close()

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		ch <- struct {
			Response []byte
			API      string
			Error    error
		}{Error: err}
		return
	}

	// Send the response back through the channel
	ch <- struct {
		Response []byte
		API      string
		Error    error
	}{
		Response: body,
		API:      apiName,
	}
}

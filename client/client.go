package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	const API_URL = "http://localhost:8080/cotacao"
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	var ctr CotacaoResponse
	err = json.Unmarshal(body, &ctr)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("Dólar:%s\n", ctr.Bid))
}

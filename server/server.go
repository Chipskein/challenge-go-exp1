package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const port = 8080

var database *sql.DB

type CotacaoAPIResponse struct {
	Usdbrl struct {
		Ask        string `json:"ask"`
		Bid        string `json:"bid"`
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		CreateDate string `json:"create_date"`
		High       string `json:"high"`
		Low        string `json:"low"`
		Name       string `json:"name"`
		PctChange  string `json:"pctChange"`
		Timestamp  string `json:"timestamp"`
		VarBid     string `json:"varBid"`
	} `json:"USDBRL"`
}
type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func SaveBid(bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	conn, err := database.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.ExecContext(ctx, "INSERT INTO bids(bid,created) values (?,datetime('now'))", bid)
	return err
}

func ContacaoHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.NotFound(w, req)
	}
	const bidurl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bidurl, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var ctr CotacaoAPIResponse
	err = json.Unmarshal(body, &ctr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = SaveBid(ctr.Usdbrl.Bid)
	if err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}
	err = json.NewEncoder(w).Encode(&CotacaoResponse{Bid: ctr.Usdbrl.Bid})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func main() {
	mux := http.NewServeMux()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS bids(bid decimal,created text);")
	if err != nil {
		log.Fatal(err)
	}
	database = db
	mux.HandleFunc("/cotacao", ContacaoHandler)
	fmt.Printf("Iniciando Servidor na porta:%d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatalf("Error:%s", err)
	}
}

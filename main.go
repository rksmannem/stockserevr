package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const destURL = "http://worldtradingdata.com/api/v1/stock"
const apiToken = "uQUdb9DzzEM6APol77hf9W9FVIuVvac9peGC8geLa6qMcJv18FdDw9NLAoKp"

// MyHandler ....
func MyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// stockExchange := r.URL.Query().Get("stock_exchange")
	// log.Println("stockExchange : ", stockExchange)

	req, err := http.NewRequest("GET", destURL, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := req.URL.Query()
	q.Add("api_token", apiToken)
	q.Add("symbol", vars["symbol"])
	req.URL.RawQuery = q.Encode()

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	response, _ := netClient.Get(req.URL.String())

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// responseString := string(responseData)
	// fmt.Fprint(w, responseString)

	stkData := decodeNTransformJSON(responseData)
	fmt.Fprint(w, string(stkData))
}

func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/stock/{symbol}", MyHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", mux))
}

type internalResp struct {
	symbol          string
	name            string
	price           string
	close_yesterday string
	currency        string
	market_cap      string
	volume          string
	timezone        string
	timezone_name   string
	gmt_offset      string
	last_trade_time string
}

// respDaataOutPut ...
type respDataOutPut struct {
	data map[string]internalResp
}

func decodeNTransformJSON(jsonData []byte) []byte {
	var out interface{}
	err := json.Unmarshal(jsonData, &out)
	if err != nil {
		log.Println(err)
	}
	tmpData := out.(map[string]interface{})

	// var respOut map[string]interface{}
	// var internalData internalResp

	// respOut := make(map[string]interface{})

	var respOut respDataOutPut
	for _, v := range tmpData {
		switch v := v.(type) {
		case []interface{}:
			for _, u := range v {
				out := u.(map[string]interface{})
				exchangeName := out["stock_exchange_short"].(string)
				// respOut[exchangeName] = internalResp{
				// 	symbol:          out["symbol"].(string),
				// 	name:            out["name"].(string),
				// 	price:           out["price"].(string),
				// 	close_yesterday: out["close_yesterday"].(string),
				// 	currency:        out["currency"].(string),
				// 	market_cap:      out["market_cap"].(string),
				// 	volume:          out["volume"].(string),
				// 	timezone:        out["timezone"].(string),
				// 	timezone_name:   out["timezone_name"].(string),
				// 	gmt_offset:      out["gmt_offset"].(string),
				// 	last_trade_time: out["last_trade_time"].(string),
				// }

				// respOut = u.(map[string]interface{})

				internalData := internalResp{
					symbol:          out["symbol"].(string),
					name:            out["name"].(string),
					price:           out["price"].(string),
					close_yesterday: out["close_yesterday"].(string),
					currency:        out["currency"].(string),
					market_cap:      out["market_cap"].(string),
					volume:          out["volume"].(string),
					timezone:        out["timezone"].(string),
					timezone_name:   out["timezone_name"].(string),
					gmt_offset:      out["gmt_offset"].(string),
					last_trade_time: out["last_trade_time"].(string),
				}
				respOut.data = map[string]internalResp{
					exchangeName: internalData,
				}
			}
		}
	}

	fmt.Println("respOut:", respOut.data)
	respData, err := json.Marshal(respOut.data)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("respData:", string(respData))
	return respData
}

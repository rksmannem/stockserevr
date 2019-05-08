package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const destURL = "http://worldtradingdata.com/api/v1/stock"
const apiToken = "uQUdb9DzzEM6APol77hf9W9FVIuVvac9peGC8geLa6qMcJv18FdDw9NLAoKp"

var jsonResponseFields = []string{"symbol", "name",
	"price", "close_yesterday", "currency", "market_cap", "volume",
	"timezone", "timezone_name", "gmt_offset", "last_trade_time"}

func stockExchngHndlr(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockExchanges := r.URL.Query().Get("stock_exchange")
	if stockExchanges == "" {
		stockExchanges = "AMEX"
	}
	exchngQueryParamsList := strings.Split(stockExchanges, ",")

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

	decodedData := decodeNTransformJSON(responseData)

	var jsonResp []byte
	out := make(map[string]map[string]interface{})
	for exchngName, resp := range decodedData {
		if Find(exchngQueryParamsList, exchngName) {
			out[exchngName] = resp
			stkData := encodeMapData(out)
			jsonResp = append(jsonResp, stkData...)
		}
	}

	if len(jsonResp) == 0 {
		jsonResp = []byte(fmt.Sprintf("symbol:%s not found in stock exchanges:%v", vars["symbol"], exchngQueryParamsList))
	}
	fmt.Fprint(w, string(jsonResp))
}

func decodeNTransformJSON(jsonData []byte) map[string]map[string]interface{} {
	var out interface{}
	err := json.Unmarshal(jsonData, &out)
	if err != nil {
		log.Fatal(err)
	}

	respOut := make(map[string]map[string]interface{})
	for _, v := range out.(map[string]interface{}) {
		switch v := v.(type) {
		case []interface{}:
			for _, u := range v {
				out := u.(map[string]interface{})

				requiredFieldsMap := make(map[string]interface{})
				for _, prop := range jsonResponseFields {
					requiredFieldsMap[prop] = out[prop]
				}

				exchangeName := out["stock_exchange_short"].(string)
				respOut = map[string]map[string]interface{}{
					exchangeName: requiredFieldsMap,
				}
			}
		}
	}
	return respOut
}

func encodeMapData(respOut map[string]map[string]interface{}) []byte {
	respData, err := json.Marshal(respOut)
	if err != nil {
		log.Fatal(err)
	}
	return respData
}

func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/stock/{symbol}", stockExchngHndlr).Methods("GET")
	mux.HandleFunc("/stock/{symbol}", stockExchngHndlr).Queries("stock_exchange", "{AMEX}").Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", mux))
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// TickerGet .
func TickerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	// api request
	req, _ := http.NewRequest("GET", "https://api.kite.trade/instruments", nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		SetReponseStatus(w, r, statusCodeBadRequest, "", dialogType, response)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	lines := strings.Split(string(body), "\r")
	nse := []map[string]string{}
	bse := []map[string]string{}
	nfo := []map[string]string{}
	lines = lines[1:]
	for _, line := range lines {
		fields := strings.Split(strings.Trim(line, ""), ",")
		if len(fields) > 4 {
			if strings.EqualFold(fields[11], "NSE") {
				nse = append(nse, map[string]string{
					"i":   fields[0],  // instrumentToken
					"e":   fields[1],  // exchangeToken
					"t":   fields[2],  // tradingSymbol
					"n":   fields[3],  // name
					"ex":  fields[5],  // expiry
					"s":   fields[6],  // strike
					"ts":  fields[7],  // tickSize
					"l":   fields[8],  // lotSize
					"it":  fields[9],  // instrumentType
					"se":  fields[10], // segment
					"exc": fields[11], // exchange
				})
			} else if strings.EqualFold(fields[11], "BSE") {
				bse = append(bse, map[string]string{
					"i":   fields[0],  // instrumentToken
					"e":   fields[1],  // exchangeToken
					"t":   fields[2],  // tradingSymbol
					"n":   fields[3],  // name
					"ex":  fields[5],  // expiry
					"s":   fields[6],  // strike
					"ts":  fields[7],  // tickSize
					"l":   fields[8],  // lotSize
					"it":  fields[9],  // instrumentType
					"se":  fields[10], // segment
					"exc": fields[11], // exchange
				})
			} else if strings.EqualFold(fields[11], "NFO") {
				nfo = append(nfo, map[string]string{
					"i":   fields[0],  // instrumentToken
					"e":   fields[1],  // exchangeToken
					"t":   fields[2],  // tradingSymbol
					"n":   fields[3],  // name
					"ex":  fields[5],  // expiry
					"s":   fields[6],  // strike
					"ts":  fields[7],  // tickSize
					"l":   fields[8],  // lotSize
					"it":  fields[9],  // instrumentType
					"se":  fields[10], // segment
					"exc": fields[11], // exchange
				})
			}
		}
	}

	tickers := []map[string]string{}
	response["data"] = append(append(append(tickers, nse...), bse...), nfo...)
	response["meta"] = setMeta(statusCodeOk, "", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

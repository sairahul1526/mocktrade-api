package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var userIDs map[string]string
var positions map[string]map[string]string
var tickerIDs []string
var tickerIDsMap map[string]string
var tickers map[string]float64
var userTickerIDs map[string]map[string]string

// daily amount

// DailyAmount .
func DailyAmount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")

	go func() {
		dailyAmountGetUserIDs()
		dailyAmountGetTickerIDs()
		dailyAmountGetUserTickerIDs()
		dailyAmountGetTickerDetails()
		dailyAmountCalculate()
	}()

	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

func dailyAmountGetUserIDs() {
	userIDs = map[string]string{}
	rows, err := db.Query("select user_id, amount from " + accountTable + "")
	if err != nil {
		return
	}

	var (
		userID string
		amount string
	)
	for rows.Next() {
		rows.Scan(&userID, &amount)
		userIDs[userID] = amount
	}
}

func dailyAmountGetTickerIDs() {
	tickerIDs = []string{}
	rows, err := db.Query("select ticker from " + positionTable + " group by ticker")
	if err != nil {
		return
	}

	var (
		tickerID string
	)
	for rows.Next() {
		rows.Scan(&tickerID)
		tickerIDs = append(tickerIDs, tickerID)
	}
}

func dailyAmountGetUserTickerIDs() {
	userTickerIDs = map[string]map[string]string{}
	rows, err := db.Query("select user_id, ticker, shares from " + positionTable + " order by user_id")
	if err != nil {
		return
	}

	var (
		userID   string
		tickerID string
		shares   string
	)
	for rows.Next() {
		rows.Scan(&userID, &tickerID, &shares)
		if userTickerIDs[userID] == nil {
			userTickerIDs[userID] = map[string]string{}
		}
		userTickerIDs[userID][tickerID] = shares
	}
}

func dailyAmountGetTickerDetails() {
	tickers = map[string]float64{}
	i := 0
	url := "https://api.kite.trade/quote/ltp?"
	init := true
	for _, ticker := range tickerIDs {
		if i > 500 {
			dailyAmountParseTickerDetails(url)
			i = 0
			url = "https://api.kite.trade/quote/ltp?"
			init = true
		} else {
			if !init {
				url += "&"
			}
			url += "i=" + ticker
			init = false
		}
	}
	dailyAmountParseTickerDetails(url)
}

func dailyAmountParseTickerDetails(url string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Kite-Version", "3")
	req.Header.Add("Authorization", "token cu50ienpvww2pb2o:titWqRrkPvkZrWt2bOjzkWmW2ouRmijX")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	tickerQuotes := TickerQuotes{}
	json.Unmarshal(body, &tickerQuotes)

	for _, ticker := range tickerIDs {
		tickers[ticker] = tickerQuotes.Data[ticker].Price
	}
}

func dailyAmountCalculate() {
	var total float64
	for userID, amount := range userIDs {
		total, _ = strconv.ParseFloat(amount, 64)
		for ticker, shares := range userTickerIDs[userID] {
			no, _ := strconv.ParseFloat(shares, 64)
			total += tickers[ticker] * no
		}
		db.Exec(buildInsertStatement(amountTable, map[string]string{
			"user_id": userID,
			"amount":  strconv.FormatFloat(total, 'f', 2, 64),
			"date":    time.Now().Format("2006-01-02"),
		}))
	}
}

// remove expired

// RemoveExpired .
func RemoveExpired(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")

	go func() {
		removeExpiredGetExpiredPostions()
		removeExpiredGetTickerDetails()
		removeExpiredCalculate()
	}()

	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

func removeExpiredGetExpiredPostions() {
	tickerIDsMap = map[string]string{}
	tickerIDs = []string{}
	positions = map[string]map[string]string{}
	rows, err := db.Query("select user_id, ticker, shares from " + positionTable + " where expiry is not null and expiry != '' and expiry < '" + time.Now().Format("2006-01-02") + "'")
	if err != nil {
		return
	}

	var (
		userID string
		ticker string
		shares string
	)
	for rows.Next() {
		rows.Scan(&userID, &ticker, &shares)
		if positions[userID] == nil {
			positions[userID] = map[string]string{}
		}
		tickerIDsMap[ticker] = "1"
		positions[userID][ticker] = shares
	}

	for key := range tickerIDsMap {
		tickerIDs = append(tickerIDs, key)
	}
}

func removeExpiredGetTickerDetails() {
	tickers = map[string]float64{}
	i := 0
	url := "https://api.kite.trade/quote/ltp?"
	init := true
	for _, ticker := range tickerIDs {
		if i > 500 {
			removeExpiredParseTickerDetails(url)
			i = 0
			url = "https://api.kite.trade/quote/ltp?"
			init = true
		} else {
			if !init {
				url += "&"
			}
			url += "i=" + ticker
			init = false
		}
	}
	removeExpiredParseTickerDetails(url)
}

func removeExpiredParseTickerDetails(url string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Kite-Version", "3")
	req.Header.Add("Authorization", "token cu50ienpvww2pb2o:XnoYIy56Yhpgf9XEfCk2FMM2jQ4LyU3L")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	tickerQuotes := TickerQuotes{}
	json.Unmarshal(body, &tickerQuotes)

	for _, ticker := range tickerIDs {
		tickers[ticker] = tickerQuotes.Data[ticker].Price
	}
}

func removeExpiredCalculate() {
	for userID, position := range positions {
		for ticker, shares := range position {
			no, _ := strconv.ParseFloat(shares, 64)
			amount := no * tickers[ticker]
			db.Exec("update " + accountTable + " set amount = amount + " + strconv.FormatFloat(amount, 'f', 2, 64) + " where user_id = '" + userID + "'")
			deleteSQL(positionTable, url.Values{"user_id": {userID}, "ticker": {ticker}})
		}
	}
}

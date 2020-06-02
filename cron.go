package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var userIDs map[string]string
var positions map[string]map[string]string
var tickersCron map[string]float64
var userTickerIDs map[string]map[string]string
var token string

// daily amount

// DailyCron .
func DailyCron(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")

	if strings.EqualFold(r.Header.Get("apikey"), cron) {
		token = r.FormValue("token")
		go func() {
			// remove expired
			removeExpiredGetExpiredPostions()
			removeExpiredCalculate()

			// daily amount
			dailyAmountGetUserIDs()
			dailyAmountGetUserTickerIDs()
			dailyAmountCalculate()
		}()
	}

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

func dailyAmountCalculate() {
	var total float64
	for userID, amount := range userIDs {
		total, _ = strconv.ParseFloat(amount, 64)
		for ticker, shares := range userTickerIDs[userID] {
			no, _ := strconv.ParseFloat(shares, 64)
			temp := strings.Split(getValueRedis(ticker), ":")
			if len(temp) > 1 {
				price, _ := strconv.ParseFloat(temp[1], 64)
				total += price * no
			}
		}
		db.Exec(buildInsertStatement(amountTable, map[string]string{
			"user_id": userID,
			"amount":  strconv.FormatFloat(total, 'f', 2, 64),
			"date":    time.Now().In(mumbai).Format("2006-01-02"),
		}))
	}
}

// remove expired

func removeExpiredGetExpiredPostions() {
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
		positions[userID][ticker] = shares
	}
}

func removeExpiredCalculate() {
	for userID, position := range positions {
		for ticker, shares := range position {
			no, _ := strconv.ParseFloat(shares, 64)
			temp := strings.Split(getValueRedis(ticker), ":")
			if len(temp) > 1 {
				price, _ := strconv.ParseFloat(temp[1], 64)
				amount := no * price
				db.Exec("update " + accountTable + " set amount = amount + " + strconv.FormatFloat(amount, 'f', 2, 64) + " where user_id = '" + userID + "'")
				deleteSQL(positionTable, url.Values{"user_id": {userID}, "ticker": {ticker}})
			}
		}
	}
}

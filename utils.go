package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
	gomail "gopkg.in/gomail.v2"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func logger(str interface{}) {
	if test {
		fmt.Println(str)
	}
}

func requiredFiledsCheck(body map[string]string, required []string) string {
	for _, field := range required {
		if len(body[field]) == 0 {
			return field
		}
	}
	return ""
}

func getTickers() []map[string]string {
	tickers := []map[string]string{}

	value := getValueRedis("tickers")
	if len(value) > 0 {
		err := json.Unmarshal([]byte(value), &tickers)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("from cache")
			return tickers
		}
	}

	kiteclient := kiteconnect.New(apiKey)
	instruments, err := kiteclient.GetInstrumentsByExchange("NSE")
	if err != nil {
		fmt.Println("getTickers", err)
		return []map[string]string{}
	}

	for _, instrument := range instruments {
		tickers = append(tickers, map[string]string{
			"i":   strconv.Itoa(instrument.InstrumentToken),
			"e":   strconv.Itoa(instrument.ExchangeToken),
			"t":   instrument.Tradingsymbol,
			"n":   instrument.Name,
			"ex":  "", // since these are stocks and no expiry
			"s":   strconv.FormatFloat(instrument.StrikePrice, 'f', 2, 64),
			"ti":  strconv.FormatFloat(instrument.TickSize, 'f', 2, 64),
			"l":   strconv.FormatFloat(instrument.LotSize, 'f', 2, 64),
			"in":  instrument.InstrumentType,
			"se":  instrument.Segment,
			"exc": instrument.Exchange,
		})
	}

	body, err := json.Marshal(tickers)
	if err != nil {
		fmt.Println(err)
	} else {
		setValueRedisWithExpiration("tickers", string(body), 15*time.Minute)
	}

	return tickers
}

func sqlErrorCheck(code uint16) string {
	if code == 1054 { // Error 1054: Unknown column
		return statusCodeBadRequest
	} else if code == 1062 { // Error 1062: Duplicate entry
		return statusCodeDuplicateEntry
	}
	return statusCodeServerError
}
func setMeta(status string, msg string, msgType string) map[string]string {
	if len(msg) == 0 {
		if status == statusCodeBadRequest {
			msg = "Bad Request Body"
		} else if status == statusCodeServerError {
			msg = "Internal Server Error"
		}
	}
	return map[string]string{
		"status":       status,
		"message":      msg,
		"message_type": msgType, // 1 : dialog or 2 : toast if msg
	}
}

func getHTTPStatusCode(code string) int {
	switch code {
	case statusCodeOk:
		return http.StatusOK
	case statusCodeCreated:
		return http.StatusCreated
	case statusCodeBadRequest:
		return http.StatusBadRequest
	case statusCodeServerError:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// SetReponseStatus .
func SetReponseStatus(w http.ResponseWriter, r *http.Request, status string, msg string, msgType string, response map[string]interface{}) {
	w.Header().Set("Status", status)
	response["meta"] = setMeta(status, msg, msgType)
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

func checkAppUpdate(r *http.Request) (map[string]string, bool) {
	if strings.EqualFold(r.Header.Get("apikey"), androidLive) || strings.EqualFold(r.Header.Get("apikey"), androidTest) {
		appversion, _ := strconv.ParseFloat(r.Header.Get("appversion"), 64)
		if appversion < androidForceVersionCode {
			return setMeta(statusCodeOk, "App update required", appUpdateRequired), true
		} else if appversion < androidVersionCode {
			return setMeta(statusCodeOk, "App update available", appUpdateAvailable), true
		}
	} else if strings.EqualFold(r.Header.Get("apikey"), iOSLive) || strings.EqualFold(r.Header.Get("apikey"), iOSTest) {
		appversion, _ := strconv.ParseFloat(r.Header.Get("appversion"), 64)
		if appversion < iOSForceVersionCode {
			return setMeta(statusCodeOk, "App update required", appUpdateRequired), true
		} else if appversion < iOSVersionCode {
			return setMeta(statusCodeOk, "App update available", appUpdateAvailable), true
		}
	}
	return map[string]string{}, false
}

func checkHeaders(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var response = make(map[string]interface{})
		if len(r.Header.Get("apikey")) == 0 || len(r.Header.Get("appversion")) == 0 {
			SetReponseStatus(w, r, statusCodeBadRequest, "apikey, appversion required", "", response)
			return
		} else if len(apikeys[r.Header.Get("apikey")]) == 0 {
			SetReponseStatus(w, r, statusCodeBadRequest, "Unauthorized request. Not valid apikey", "", response)
			return
		}

		if migrate { // statusCodeBadRequest because app will hit again if 500
			SetReponseStatus(w, r, statusCodeBadRequest, "Server is busy. Please try after some time", dialogType, response)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// RandStringBytes .
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func checkUserID(userID string) bool {
	var status string
	db.QueryRow("select status from " + accountTable + " where user_id = '" + userID + "' and status = 1").Scan(&status)
	return len(status) > 0
}

func isMarketOpen() bool {
	now := time.Now().UTC()
	var (
		openTime  string
		closeTime string
		holiday   string
	)
	db.QueryRow("select open, close, holiday from "+timingTable+" where day = '"+strconv.Itoa(int(now.Weekday()))+"'").Scan(&openTime, &closeTime, &holiday)

	if len(openTime) == 0 {
		return false
	}
	if strings.EqualFold(holiday, "1") {
		return false
	}
	openArr := strings.Split(openTime, ":")
	closeArr := strings.Split(closeTime, ":")
	hour, _ := strconv.Atoi(openArr[0])
	min, _ := strconv.Atoi(openArr[1])
	open := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.UTC).Add(-330 * time.Minute)
	if time.Now().UTC().Before(open) {
		return false
	}

	hour, _ = strconv.Atoi(closeArr[0])
	min, _ = strconv.Atoi(closeArr[1])
	close := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.UTC).Add(-330 * time.Minute)
	if time.Now().UTC().After(close) {
		return false
	}
	return true
}

func setValueRedis(key, value string) {
	redisClient.Set(ctx, key, value, 0)
}

func setValueRedisWithExpiration(key, value string, d time.Duration) {
	redisClient.Set(ctx, key, value, d)
}

func setValuesRedis(keyValues map[string]string) {
	redisClient.MSet(ctx, keyValues, 0)
}

func getValueRedis(key string) string {
	val, _ := redisClient.Get(ctx, key).Result()
	return val
}

func getValuesRedis(keys []string) []interface{} {
	val, _ := redisClient.MGet(ctx, keys...).Result()
	return val
}

func msg91(to, otp string) {
	url := "https://api.msg91.com/api/v2/sendsms"

	payload := strings.NewReader("{ \"sender\": \"MOCKAP\", \"route\": \"4\", \"country\": \"91\", \"sms\": [ { \"message\": \"" + strings.ReplaceAll(otpMessage, "##OTP##", otp) + "\", \"to\": [ \"" + to + "\"] } ] }")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("authkey", "331064AVVjQRraN7jt5ed65c82P1")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("SendOTP", err)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
}

// defer measureTime("expensivePrint")()
// To log code latency
func measureTime(funcName string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("Time taken by %s function is %v \n", funcName, time.Since(start))
	}
}

func mail(to, title, body string) {
	fmt.Println("mail", to, title, body)
	m := gomail.NewMessage()
	m.SetHeader("From", "rahul.mocktrade@gmail.com")
	m.SetHeader("To", "rahul.mocktrade@gmail.com")

	m.SetHeader("Bcc", to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "rahul.mocktrade@gmail.com", "Gmail9848$")

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("mail", to, err)
	}
}

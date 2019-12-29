package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func isMarketOpen() bool {
	now := time.Now()
	var (
		openTime  string
		closeTime string
		holiday   string
	)
	db.QueryRow("select open, close, holiday from "+timingTable+" where day = '"+openTime+"'").Scan(&openTime, &closeTime, &holiday)
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
	open := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.Local).Add(-330 * time.Minute)
	if time.Since(open) < 0 {
		return false
	}

	hour, _ = strconv.Atoi(closeArr[0])
	min, _ = strconv.Atoi(closeArr[1])
	close := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.Local).Add(-330 * time.Minute)
	if time.Since(close) > 0 {
		return false
	}
	return true
}

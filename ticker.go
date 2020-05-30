package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// TickerGet .
func TickerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["data"] = getTickers()
	response["meta"] = setMeta(statusCodeOk, "", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

// TickerCloseGet .
func TickerCloseGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	data := []map[string]string{}
	tokensClose := []string{}
	for _, token := range tokens {
		tokensClose = append(tokensClose, strconv.Itoa(int(token))+"_close")
	}
	closes := getValuesRedis(tokensClose)
	temp := []string{}
	for _, close := range closes {
		if close != nil {
			temp = strings.Split(close.(string), ":")
			if len(temp) > 1 {
				data = append(data, map[string]string{
					"k": temp[0],
					"v": close.(string),
				})
			}
		}
	}
	response["data"] = data
	response["meta"] = setMeta(statusCodeOk, "", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	for _, token := range tokens {
		data = append(data, map[string]string{
			"k": strconv.Itoa(int(token)),
			"v": getValueRedis(strconv.Itoa(int(token)) + "_close"),
		})
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

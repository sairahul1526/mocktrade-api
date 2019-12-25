package main

import (
	"encoding/json"
	"net/http"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

// Login .
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")
	kc := kiteconnect.New(apiKey)
	response["data"] = []map[string]string{
		map[string]string{
			"url": kc.GetLoginURL(),
		},
	}
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

package main

import (
	"encoding/json"
	"net/http"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

// Token .
func Token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")

	kc := kiteconnect.New(apiKey)
	requestToken := r.FormValue("tok")
	data, err := kc.GenerateSession(requestToken, apiSecret)
	if err != nil {
		return
	}

	response["data"] = []map[string]string{
		map[string]string{
			"userid": data.UserID,
			"token":  data.AccessToken,
		},
	}
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

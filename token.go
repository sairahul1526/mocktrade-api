package main

import (
	"encoding/json"
	"net/http"
	"os"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

// Token .
func Token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})
	response["meta"] = setMeta(statusCodeOk, "ok", "")

	kc := kiteconnect.New(apiKey)
	requestToken := r.FormValue("token")
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

// TokenUpdate .
func TokenUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	os.Setenv("accessToken", r.FormValue("token"))
	os.Exit(3)

	response["meta"] = setMeta(statusCodeOk, "ok", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

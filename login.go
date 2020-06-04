package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
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

// SendOTP .
func SendOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, []string{"phone"})
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}
	if len(body["phone"]) != 10 {
		SetReponseStatus(w, r, statusCodeBadRequest, "Valid Phone Number Required", dialogType, response)
		return
	}

	go func(r *http.Request) {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "mock.trade",
			AccountName: "mock@trade",
			Secret:      []byte(body["phone"]),
		})
		if err != nil {
			fmt.Println(err)
		}
		otp, err := totp.GenerateCodeCustom(key.Secret(), time.Now().In(mumbai), totp.ValidateOpts{Period: 300, Digits: 6, Skew: 1})
		if err != nil {
			fmt.Println(err)
		}

		msg91(body["phone"], otp)

	}(r)

	w.Header().Set("Status", statusCodeOk)
	response["meta"] = setMeta(statusCodeOk, "", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

// VerifyOTP .
func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, []string{"phone", "otp", "user_id"})
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "mock.trade",
		AccountName: "mock@trade",
		Secret:      []byte(body["phone"]),
	})
	if err != nil {
		fmt.Println(err)
	}
	valid, err := totp.ValidateCustom(body["otp"], key.Secret(), time.Now().In(mumbai), totp.ValidateOpts{Period: 300, Digits: 6, Skew: 1})
	if err != nil {
		fmt.Println(err)
	}
	if valid {
		data, status, ok := selectProcess("select * from " + accountTable + " where phone = '" + body["phone"] + "' and status != 3")
		w.Header().Set("Status", status)
		if ok {
			if len(data) > 0 {
				response["data"] = data
			} else {
				updateSQL(accountTable, url.Values{"user_id": {body["user_id"]}}, map[string]string{
					"phone":              body["phone"],
					"modified_date_time": time.Now().In(mumbai).String(),
				})
			}
		} else {
			SetReponseStatus(w, r, status, "", dialogType, response)
			return
		}
		response["meta"] = setMeta(statusCodeOk, "", "")
	} else {
		w.Header().Set("Status", statusCodeBadRequest)
		response["meta"] = setMeta(statusCodeBadRequest, "Wrong OTP", dialogType)
	}
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

// SendEmailOTP .
func SendEmailOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, []string{"email"})
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}

	go func(r *http.Request) {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "mock.trade",
			AccountName: "mock@trade",
			Secret:      []byte(body["email"]),
		})
		if err != nil {
			fmt.Println(err)
		}
		otp, err := totp.GenerateCodeCustom(key.Secret(), time.Now().In(mumbai), totp.ValidateOpts{Period: 300, Digits: 6, Skew: 1})
		if err != nil {
			fmt.Println(err)
		}

		mail(body["email"], strings.Replace(otpEmailSubject, "##OTP##", otp, -1), strings.Replace(otpEmailBody, "##OTP##", otp, -1))
	}(r)

	w.Header().Set("Status", statusCodeOk)
	response["meta"] = setMeta(statusCodeOk, "", "")
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

// VerifyEmailOTP .
func VerifyEmailOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, []string{"email", "otp", "user_id"})
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "mock.trade",
		AccountName: "mock@trade",
		Secret:      []byte(body["email"]),
	})
	if err != nil {
		fmt.Println(err)
	}
	valid, err := totp.ValidateCustom(body["otp"], key.Secret(), time.Now().In(mumbai), totp.ValidateOpts{Period: 300, Digits: 6, Skew: 1})
	if err != nil {
		fmt.Println(err)
	}
	if valid {
		data, status, ok := selectProcess("select * from " + accountTable + " where email = '" + body["email"] + "' and status != 3")
		w.Header().Set("Status", status)
		if ok {
			if len(data) > 0 {
				response["data"] = data
			} else {
				updateSQL(accountTable, url.Values{"user_id": {body["user_id"]}}, map[string]string{
					"email":              body["email"],
					"modified_date_time": time.Now().In(mumbai).String(),
				})
			}
		} else {
			SetReponseStatus(w, r, status, "", dialogType, response)
			return
		}
		response["meta"] = setMeta(statusCodeOk, "", "")
	} else {
		w.Header().Set("Status", statusCodeBadRequest)
		response["meta"] = setMeta(statusCodeBadRequest, "Wrong OTP", dialogType)
	}
	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	json.NewEncoder(w).Encode(response)
}

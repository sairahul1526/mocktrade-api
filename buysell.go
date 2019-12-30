package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BuySellAdd .
func BuySellAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, buySellRequiredFields)
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}

	if !isMarketOpen() {
		SetReponseStatus(w, r, statusCodeBadRequest, "Order was placed outside of trading hours.", dialogType, response)
		return
	}

	body["status"] = "1"
	body["created_date_time"] = time.Now().In(mumbai).String()

	tx, err := db.Begin()
	if err != nil {
		SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
		return
	}
	if strings.EqualFold(body["type"], "1") {
		amountData, _, _ := selectProcess("select amount from " + accountTable + " where user_id = '" + body["user_id"] + "'")
		amount, _ := strconv.ParseFloat(amountData[0]["amount"], 64)
		invested, _ := strconv.ParseFloat(body["invested"], 64)
		if amount >= invested {
			_, err = tx.Exec(buildInsertStatement(positionTable, map[string]string{
				"user_id":           body["user_id"],
				"ticker":            body["ticker"],
				"name":              body["name"],
				"invested":          body["invested"],
				"shares":            body["shares"],
				"status":            "1",
				"expiry":            body["expiry"],
				"created_date_time": body["created_date_time"],
			}) + " on duplicate key update invested = invested + " + body["invested"] +
				", shares = shares + " + body["shares"] + ", modified_date_time = '" + body["created_date_time"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			_, err = tx.Exec("update " + accountTable + " set amount = amount - " + body["invested"] + ", modified_date_time = '" + body["created_date_time"] + "' where user_id = '" + body["user_id"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			delete(body, "expiry")
			_, err = tx.Exec(buildInsertStatement(orderTable, body))
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			response["meta"] = setMeta(statusCodeOk, "Order complete", "")
		} else {
			response["meta"] = setMeta(statusCodeBadRequest, "Insufficient funds. Amount required is "+body["invested"]+", but available amount is "+amountData[0]["amount"], dialogType)
		}
	} else {
		positionData, _, _ := selectProcess("select shares from " + positionTable + " where user_id = '" + body["user_id"] + "' and ticker = '" + body["ticker"] + "'")
		sharesAvailable, _ := strconv.ParseFloat(positionData[0]["shares"], 64)
		sharesToSell, _ := strconv.ParseFloat(body["shares"], 64)
		if sharesAvailable > sharesToSell {
			_, err = tx.Exec(buildInsertStatement(positionTable, map[string]string{
				"user_id":           body["user_id"],
				"ticker":            body["ticker"],
				"name":              body["name"],
				"invested":          body["invested"],
				"shares":            body["shares"],
				"status":            "1",
				"expiry":            body["expiry"],
				"created_date_time": body["created_date_time"],
			}) + " on duplicate key update invested = invested - " + body["invested"] +
				", shares = shares - " + body["shares"] + ", modified_date_time = '" + body["created_date_time"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			_, err = tx.Exec("update " + accountTable + " set amount = amount + " + body["invested"] + ", modified_date_time = '" + body["created_date_time"] + "' where user_id = '" + body["user_id"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			delete(body, "expiry")
			_, err = tx.Exec(buildInsertStatement(orderTable, body))
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			response["meta"] = setMeta(statusCodeOk, "Order complete", "")
		} else if sharesAvailable == sharesToSell {
			_, err = tx.Exec("delete from " + positionTable + " where user_id = '" + body["user_id"] + "' and ticker = '" + body["ticker"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			_, err = tx.Exec("update " + accountTable + " set amount = amount + " + body["invested"] + ", modified_date_time = '" + body["created_date_time"] + "' where user_id = '" + body["user_id"] + "'")
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			delete(body, "expiry")
			_, err = tx.Exec(buildInsertStatement(orderTable, body))
			if err != nil {
				tx.Rollback()
				SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
				return
			}
			response["meta"] = setMeta(statusCodeOk, "Order complete", "")
		} else {
			response["meta"] = setMeta(statusCodeBadRequest, "Insufficient holdings. Only "+positionData[0]["shares"]+" available to sell", dialogType)
		}
	}

	err = tx.Commit()
	if err != nil {
		SetReponseStatus(w, r, statusCodeBadRequest, "Order not placed", dialogType, response)
		return
	}

	w.Header().Set("Status", response["meta"].(map[string]string)["status"])

	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

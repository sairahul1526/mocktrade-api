package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// AlertGet .
func AlertGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	params := r.URL.Query()
	limitOffset := " "

	if _, ok := params["limit"]; ok {
		limitOffset += " limit " + params["limit"][0]
		delete(params, "limit")
	} else {
		limitOffset += " limit " + defaultLimit
	}

	offset := defaultOffset
	if _, ok := params["offset"]; ok {
		limitOffset += " offset " + params["offset"][0]
		offset = params["offset"][0]
		delete(params, "offset")
	} else {
		limitOffset += " offset " + defaultOffset
	}

	orderBy := " "

	if _, ok := params["orderby"]; ok {
		orderBy += " order by " + params["orderby"][0]
		delete(params, "orderby")
		if _, ok := params["sortby"]; ok {
			orderBy += " " + params["sortby"][0] + " "
			delete(params, "sortby")
		} else {
			orderBy += " asc "
		}
	}

	resp := " * "
	if _, ok := params["resp"]; ok {
		resp = " " + params["resp"][0] + " "
		delete(params, "resp")
	}

	where := ""
	init := false
	for key, val := range params {
		if init {
			where += " and "
		}
		where += " `" + key + "` = '" + val[0] + "' "
		init = true
	}
	SQLQuery := " from `" + alertTable + "`"
	if strings.Compare(where, "") != 0 {
		SQLQuery += " where " + where
	}
	SQLQuery += orderBy
	SQLQuery += limitOffset

	data, status, ok := selectProcess("select " + resp + SQLQuery)
	w.Header().Set("Status", status)
	if ok {
		response["data"] = data

		pagination := map[string]string{}
		if len(where) > 0 {
			count, _, _ := selectProcess("select count(*) as ctn from `" + alertTable + "` where " + where)
			pagination["total_count"] = count[0]["ctn"]
		} else {
			count, _, _ := selectProcess("select count(*) as ctn from `" + alertTable + "`")
			pagination["total_count"] = count[0]["ctn"]
		}
		pagination["count"] = strconv.Itoa(len(data))
		pagination["offset"] = offset
		response["pagination"] = pagination

		response["meta"] = setMeta(status, "ok", "")
	} else {
		response["meta"] = setMeta(status, "", dialogType)
	}

	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

// AlertAdd .
func AlertAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response = make(map[string]interface{})

	body := map[string]string{}

	r.ParseMultipartForm(32 << 20)

	for key, value := range r.Form {
		body[key] = value[0]
	}
	fieldCheck := requiredFiledsCheck(body, alertRequiredFields)
	if len(fieldCheck) > 0 {
		SetReponseStatus(w, r, statusCodeBadRequest, fieldCheck+" required", dialogType, response)
		return
	}

	delete(body, "expiry")
	body["status"] = "1"
	body["created_date_time"] = time.Now().In(mumbai).String()

	id, status, ok := insertSQLWithID(alertTable, body)
	w.Header().Set("Status", status)
	if ok {
		price, _ := strconv.ParseFloat(body["price"], 64)
		redisClient.ZAdd(ctx, body["ticker"]+"_"+body["when"], &redis.Z{Score: price, Member: strconv.Itoa(int(id))})
		response["id"] = strconv.Itoa(int(id))
		if strings.EqualFold(body["when"], "0") {
			response["meta"] = setMeta(status, "You will be notified when the price of "+body["name"]+" moves below "+body["price"], dialogType)
		} else {
			response["meta"] = setMeta(status, "You will be notified when the price of "+body["name"]+" moves above "+body["price"], dialogType)
		}
	} else {
		response["meta"] = setMeta(status, "", dialogType)
	}

	w.WriteHeader(getHTTPStatusCode(response["meta"].(map[string]string)["status"]))
	meta, required := checkAppUpdate(r)
	if required {
		response["meta"] = meta
	}
	json.NewEncoder(w).Encode(response)
}

func loadAlerts() {
	alerts, status, ok := selectProcess("select * from " + alertTable + " where status = 1 and alerted = 0")
	if !ok {
		fmt.Println("loadAlerts", status)
		return
	}

	for _, alert := range alerts {
		price, _ := strconv.ParseFloat(alert["price"], 64)
		redisClient.ZAdd(ctx, alert["ticker"]+"_"+alert["when"], &redis.Z{Score: price, Member: alert["id"]})
	}
}

func alerting(ticker, price string) {
	// less than : when = 0
	result := redisClient.ZRangeByScore(ctx, ticker+"_0", &redis.ZRangeBy{Min: price, Max: "1000000000"})
	if len(result.Val()) > 0 {
		for _, id := range result.Val() {
			redisClient.SAdd(ctx, "alertIDs", id)
		}
		redisClient.ZRem(ctx, ticker+"_0", result.Val())
	}

	// greater than : when = 1
	result = redisClient.ZRangeByScore(ctx, ticker+"_1", &redis.ZRangeBy{Min: "0", Max: price})
	if len(result.Val()) > 0 {
		for _, id := range result.Val() {
			redisClient.SAdd(ctx, "alertIDs", id)
		}
		redisClient.ZRem(ctx, ticker+"_1", result.Val())
	}
}

func alertPassed() {
	ticker := time.NewTicker(time.Second)
	defer func() {
		alertPassed()
	}()
	for {
		select {
		case t := <-ticker.C:
			result := redisClient.SMembers(ctx, "alertIDs")
			if len(result.Val()) > 0 {
				db.Exec("update " + alertTable + " set alerted = 1, modified_date_time = '" + time.Now().In(mumbai).String() + "' where id in ('" + strings.Join(result.Val(), "','") + "')")
				redisClient.SRem(ctx, "alertIDs", result.Val())
				if false {
					fmt.Println(t)
				}
			}
		}
	}
}

func sendingAlertNotifications() {
	defer func() {
		sendingAlertNotifications()
	}()
	for {
		alertingIDs := []string{}
		alerts, status, ok := selectProcess("select * from " + alertTable + " where alerted = 1 and status = 1")
		if !ok {
			fmt.Println("sendingAlertNotifications", status)
			return
		}
		for _, alert := range alerts {
			alertingIDs = append(alertingIDs, alert["id"])
			if strings.EqualFold(alert["when"], "1") {
				go sendNotifications(alert["user_id"], "Alert Triggered", "Your alert of "+alert["name"]+" price greater than "+alert["price"]+" is triggered")
			} else if strings.EqualFold(alert["when"], "0") {
				go sendNotifications(alert["user_id"], "Alert Triggered", "Your alert of "+alert["name"]+" price less than "+alert["price"]+" is triggered")
			}
		}
		if len(alertingIDs) > 0 {
			db.Exec("update " + alertTable + " set alerted = 2, modified_date_time = '" + time.Now().In(mumbai).String() + "' where id in ('" + strings.Join(alertingIDs, "','") + "')")
			alertingIDs = []string{}
		}
	}
}

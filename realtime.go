package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("wsHandler", err)
		return
	}
	if !checkUserID(r.FormValue("user_id")) {
		ws.Close()
		fmt.Println("wsHandler", "User not found", r.FormValue("user_id"))
		return
	}
	db.Exec("update " + accountTable + " set websocket_in_date_time = '" + time.Now().In(mumbai).String() + "', websocket_times = websocket_times + 1 where user_id = '" + r.FormValue("user_id") + "'")
	fmt.Println(r.FormValue("tickers"))

	stocks := strings.Split(r.FormValue("tickers"), ",")

	fmt.Println(stocks)

	ticker := time.NewTicker(time.Second)
	defer func() {
		db.Exec("update " + accountTable + " set websocket_out_date_time = '" + time.Now().In(mumbai).String() + "' where user_id = '" + r.FormValue("user_id") + "'")
		ws.Close()
		ticker.Stop()
	}()

	if len(r.FormValue("tickers")) > 0 {
		var send strings.Builder
		active := true
		for {
			if active {
				select {
				case t := <-ticker.C:
					for _, val := range getValuesRedis(stocks) {
						if val != nil {
							send.WriteString(val.(string) + "#")
						}
					}
					err := ws.WriteMessage(websocket.TextMessage, []byte(send.String()))
					if err != nil {
						fmt.Println("realtime", t, err)
						active = false
					}
					send.Reset()
				}
			} else {
				break
			}
		}
	}
	fmt.Println("eneded")
}

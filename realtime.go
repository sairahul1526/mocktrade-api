package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(r.FormValue("tickers"))

	stocks := strings.Split(r.FormValue("tickers"), ",")

	for i := 0; i < len(stocks); i++ {
		stocks[i] = stocks[i] + "_last"
	}

	fmt.Println(stocks)

	ticker := time.NewTicker(time.Second)
	defer func() {
		ws.Close()
		ticker.Stop()
	}()

	if len(r.FormValue("tickers")) > 0 {
		var send string
		active := true
		for {
			if active {
				select {
				case t := <-ticker.C:
					for _, val := range getValuesRedis(stocks) {
						send += val.(string) + "#"
					}
					err := ws.WriteMessage(websocket.TextMessage, []byte(send))
					if err != nil {
						fmt.Println("realtime", t, err)
						active = false
					}
					send = ""
				}
			} else {
				break
			}
		}
	}
	fmt.Println("eneded")
}

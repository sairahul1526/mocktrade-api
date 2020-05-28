package main

import (
	"fmt"
	"strconv"
	"time"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
	"github.com/zerodhatech/gokiteconnect/ticker"
)

var (
	ticker *kiteticker.Ticker
)

var tokens = []uint32{}

// Triggered when any error is raised
func onError(err error) {
	fmt.Println("Error: ", err)
}

// Triggered when websocket connection is closed
func onClose(code int, reason string) {
	fmt.Println("Close: ", code, reason)
}

// Triggered when connection is established and ready to send and accept data
func onConnect() {
	fmt.Println("Connected")
	err := ticker.Subscribe(tokens)
	if err != nil {
		fmt.Println("onConnect", err)
	}
}

// Triggered when tick is recevived
func onTick(tick kiteticker.Tick) {
	setValueRedis(strconv.Itoa(int(tick.InstrumentToken))+"_close", strconv.Itoa(int(tick.InstrumentToken))+":"+strconv.FormatFloat(tick.OHLC.Close, 'f', -1, 64))
	setValueRedis(strconv.Itoa(int(tick.InstrumentToken))+"_last", strconv.Itoa(int(tick.InstrumentToken))+":"+strconv.FormatFloat(tick.LastPrice, 'f', -1, 64))
}

// Triggered when reconnection is attempted which is enabled by default
func onReconnect(attempt int, delay time.Duration) {
	mail("dravid.rahul1526@gmail.com", "Kite Reconnect - Mock Trade", "")
	fmt.Println("Reconnect attempt ", attempt, " in ", delay.Seconds())
}

// Triggered when maximum number of reconnect attempt is made and the program is terminated
func onNoReconnect(attempt int) {
	fmt.Println("Maximum no of reconnect attempt reached: ", attempt)
}

func connectToKite() {
	ticker = kiteticker.New(apiKey, accessToken)

	loadTickersToSubscribe()

	// Assign callbacks
	ticker.OnError(onError)
	ticker.OnClose(onClose)
	ticker.OnConnect(onConnect)
	ticker.OnReconnect(onReconnect)
	ticker.OnNoReconnect(onNoReconnect)
	ticker.OnTick(onTick)

	// Start the connection
	ticker.Serve()
}

func loadTickersToSubscribe() {
	// pass tickers
	kiteclient := kiteconnect.New(apiKey)
	instruments, err := kiteclient.GetInstrumentsByExchange("NSE")
	if err != nil {
		fmt.Println("parseTokens", err)
		return
	}
	for _, instrument := range instruments {
		fmt.Println(instrument.InstrumentToken, instrument.Tradingsymbol)
		tokens = append(tokens, uint32(instrument.InstrumentToken))
	}

}

package main

import (
	"fmt"
	"strconv"
	"time"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
	"github.com/zerodhatech/gokiteconnect/ticker"
)

var (
	ticker  *kiteticker.Ticker
	ticker2 *kiteticker.Ticker
	ticker3 *kiteticker.Ticker
)

var tokens = []uint32{}
var tokens2 = []uint32{}
var tokens3 = []uint32{}

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

func onConnect2() {
	fmt.Println("Connected")
	err := ticker2.Subscribe(tokens2)
	if err != nil {
		fmt.Println("onConnect", err)
	}
}

func onConnect3() {
	fmt.Println("Connected")
	err := ticker3.Subscribe(tokens3)
	if err != nil {
		fmt.Println("onConnect", err)
	}
}

// Triggered when tick is recevived
func onTick(tick kiteticker.Tick) {
	go setValueRedis(strconv.Itoa(int(tick.InstrumentToken)), strconv.Itoa(int(tick.InstrumentToken))+":"+strconv.FormatFloat(tick.LastPrice, 'f', -1, 64)+":"+strconv.FormatFloat(tick.OHLC.Close, 'f', -1, 64))
	go alerting(strconv.Itoa(int(tick.InstrumentToken)), strconv.FormatFloat(tick.LastPrice, 'f', -1, 64))
}

// Triggered when reconnection is attempted which is enabled by default
func onReconnect(attempt int, delay time.Duration) {
	mail("dravid.rahul1526@gmail.com", "Kite Reconnect - Mock Trade", "attempt - "+strconv.Itoa(attempt))
	fmt.Println("Reconnect attempt ", attempt, " in ", delay.Seconds())
}

// Triggered when maximum number of reconnect attempt is made and the program is terminated
func onNoReconnect(attempt int) {
	fmt.Println("Maximum no of reconnect attempt reached: ", attempt)
}

func connectToKite() {
	ticker = kiteticker.New(apiKey, accessToken)
	ticker2 = kiteticker.New(apiKey, accessToken)
	ticker3 = kiteticker.New(apiKey, accessToken)

	loadTickersToSubscribe()

	// Assign callbacks
	ticker.OnError(onError)
	ticker.OnClose(onClose)
	ticker.OnConnect(onConnect)
	ticker.OnReconnect(onReconnect)
	ticker.OnNoReconnect(onNoReconnect)
	ticker.OnTick(onTick)

	// Assign callbacks
	ticker2.OnError(onError)
	ticker2.OnClose(onClose)
	ticker2.OnConnect(onConnect2)
	ticker2.OnReconnect(onReconnect)
	ticker2.OnNoReconnect(onNoReconnect)
	ticker2.OnTick(onTick)

	// Assign callbacks
	ticker3.OnError(onError)
	ticker3.OnClose(onClose)
	ticker3.OnConnect(onConnect3)
	ticker3.OnReconnect(onReconnect)
	ticker3.OnNoReconnect(onNoReconnect)
	ticker3.OnTick(onTick)

	// Start the connection
	go ticker.Serve()
	go ticker2.Serve()
	go ticker3.Serve()
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
		if len(tokens) < 3000 {
			tokens = append(tokens, uint32(instrument.InstrumentToken))
		} else if len(tokens) < 6000 {
			tokens2 = append(tokens2, uint32(instrument.InstrumentToken))
		} else if len(tokens) < 9000 {
			tokens3 = append(tokens3, uint32(instrument.InstrumentToken))
		}
	}

}

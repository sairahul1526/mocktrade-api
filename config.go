package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
)

// kite keys
var apiKey = "cu50ienpvww2pb2o"
var apiSecret = "4jhjifghp0dt4gb89e1ggkyun6i6akls"
var accessToken = "tr2qqcH1A5bZ2spZ0uzgU2Uz09vnUfMu"

// config
var dbConfig string
var connectionPool = 10
var test = true
var migrate = false
var onesignalAppID string

// tables
var alertTable = "alerts"
var amountTable = "amounts"
var accountTable = "accounts"
var timingTable = "timings"
var tokenTable = "tokens"
var orderTable = "orders"
var positionTable = "positions"

var defaultOffset = "0"
var defaultLimit = "25"

// message types
var dialogType = "1"
var toastType = "2"
var appUpdateAvailable = "3"
var appUpdateRequired = "4"

// api keys
var androidLive = "T9h9P6j2N6y9M3Q8"
var androidTest = "K7b3V4h3C7t6g6M7"
var iOSLive = "b4E6U9K8j6b5E9W3"
var iOSTest = "R4n7N8G4m9B4S5n2"
var cron = "ZNPZTTDEVAStYczW"

// app update versions
var iOSVersionCode = 1.0
var iOSForceVersionCode = 1.0

var androidVersionCode = 1.6
var androidForceVersionCode = 1.8

// for checking unauth request
var apikeys = map[string]string{
	androidLive: "1", // android live
	androidTest: "1", // android test
	iOSLive:     "1", // iOS live
	iOSTest:     "1", // iOS test
	cron:        "1", // cron
}

// server codes
var statusCodeOk = "200"
var statusCodeCreated = "201"
var statusCodeBadRequest = "400"
var statusCodeForbidden = "403"
var statusCodeServerError = "500"
var statusCodeDuplicateEntry = "1000"

// required fields
var alertRequiredFields = []string{"user_id", "ticker", "name", "price", "when"}
var amountRequiredFields = []string{"user_id", "amount", "date"}
var accountRequiredFields = []string{}
var timingRequiredFields = []string{"day", "holiday", "open", "close"}
var tokenRequiredFields = []string{}
var orderRequiredFields = []string{"user_id", "ticker", "price", "shares", "type"}
var positionRequiredFields = []string{"user_id", "ticker", "invested", "shares"}
var buySellRequiredFields = []string{"user_id", "ticker", "name", "exchange", "price", "shares", "type"}

var otpMessage = "<#> ##OTP## is the OTP for your MockTrade App login. Valid for 60 sec. "
var otpEmailSubject = "MockTrade - Verification code: ##OTP##"
var otpEmailBody = "Hi,<br><br>Please enter below verification code to verify your email.<br><br> <h2>##OTP##</h2> <br><br>Best Regards,<br> MockTrade Team"

var mumbai *time.Location

var upgrader = websocket.Upgrader{
	WriteBufferSize:   512,
	ReadBufferSize:    0,
	WriteBufferPool:   &sync.Pool{},
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

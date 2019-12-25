package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var err error

// HealthCheck .
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("ok")
}

func connectDatabase() {
	db, err = sql.Open("mysql", dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(connectionPool)
	db.SetMaxIdleConns(connectionPool)
	db.SetConnMaxLifetime(time.Hour)
}

func inits() {

	rand.Seed(time.Now().UnixNano())

	connectDatabase()
}

func main() {

	if len(os.Getenv("dbConfig")) > 0 {
		dbConfig = os.Getenv("dbConfig")
	}
	if len(os.Getenv("connectionPool")) > 0 {
		connectionPool, _ = strconv.Atoi(os.Getenv("connectionPool"))
	}
	if len(os.Getenv("test")) > 0 {
		test, _ = strconv.ParseBool(os.Getenv("test"))
	}
	if len(os.Getenv("migrate")) > 0 {
		migrate, _ = strconv.ParseBool(os.Getenv("migrate"))
	}

	inits()
	defer db.Close()
	router := mux.NewRouter()

	// cron
	router.Path("/dailycron").HandlerFunc(checkHeaders(DailyCron)).Methods("GET")

	router.Path("/account").HandlerFunc(checkHeaders(AccountGet)).Methods("GET")
	router.Path("/account").HandlerFunc(checkHeaders(AccountAdd)).Methods("POST")
	router.Path("/account").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(AccountUpdate)).Methods("PUT")

	router.Path("/buysell").HandlerFunc(checkHeaders(BuySellAdd)).Methods("POST")

	router.Path("/timing").HandlerFunc(TimingGet).Methods("GET")
	router.Path("/timing").HandlerFunc(checkHeaders(TimingAdd)).Methods("POST")
	router.Path("/timing").Queries(
		"day", "{day}",
	).HandlerFunc(checkHeaders(TimingUpdate)).Methods("PUT")

	router.Path("/order").HandlerFunc(checkHeaders(OrderGet)).Methods("GET")
	router.Path("/order").HandlerFunc(checkHeaders(OrderAdd)).Methods("POST")
	router.Path("/order").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(OrderUpdate)).Methods("PUT")

	router.Path("/position").HandlerFunc(checkHeaders(PositionGet)).Methods("GET")
	router.Path("/position").HandlerFunc(checkHeaders(PositionAdd)).Methods("POST")
	router.Path("/position").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(PositionUpdate)).Methods("PUT")

	router.Path("/ticker").HandlerFunc(TickerGet).Methods("GET")

	router.Path("/login").HandlerFunc(Login).Methods("GET")
	router.Path("/token").HandlerFunc(Token).Methods("GET")
	router.Path("/").HandlerFunc(HealthCheck).Methods("GET")

	fmt.Println(http.ListenAndServe(":5000", &WithCORS{router}))
}

// WithCORS .
type WithCORS struct {
	r *mux.Router
}

func (s *WithCORS) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT,DELETE")
	res.Header().Set("Access-Control-Allow-Headers", "Content-Type,api_key,appversion")

	// Stop here for a Preflighted OPTIONS request.
	if req.Method == "OPTIONS" {
		return
	}

	s.r.ServeHTTP(res, req)
}

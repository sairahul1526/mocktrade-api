package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql" // for mysql
	"github.com/gorilla/mux"
)

var db *sql.DB
var err error

var ctx = context.Background()
var redisClient *redis.Client

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

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	ctx = redisClient.Context()
	pong, err := redisClient.Ping(ctx).Result()
	fmt.Println(pong, err)

}

func inits() {
	rand.Seed(time.Now().UnixNano())
	connectDatabase()
}

func main() {

	mumbai, _ = time.LoadLocation("Asia/Kolkata")
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

	// defer profile.Start().Stop()
	// defer profile.Start(profile.MemProfile).Stop()
	// go tool pprof --pdf main /var/folders/k6/0m4k5qg110jgfzpfdrdjv1dr0000gn/T/profile113169278/cpu.pprof > pprofs/file17.pdf

	inits()
	defer db.Close()
	defer redisClient.Close()
	router := mux.NewRouter()

	if len(getValueRedis("accessToken")) > 0 {
		accessToken = getValueRedis("accessToken")
	}

	go connectToKite()

	// cron
	router.Path("/dailycron").HandlerFunc(checkHeaders(DailyCron)).Methods("GET")

	router.Path("/account").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(AccountGet)).Methods("GET")
	router.Path("/account").HandlerFunc(checkHeaders(AccountAdd)).Methods("POST")
	router.Path("/account").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(AccountUpdate)).Methods("PUT")

	router.Path("/amount").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(AmountGet)).Methods("GET")

	router.Path("/buysell").HandlerFunc(checkHeaders(BuySellAdd)).Methods("POST")

	router.Path("/timing").HandlerFunc(TimingGet).Methods("GET")

	router.Path("/order").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(OrderGet)).Methods("GET")

	router.Path("/position").Queries(
		"user_id", "{user_id}",
	).HandlerFunc(checkHeaders(PositionGet)).Methods("GET")

	router.Path("/ticker").HandlerFunc(checkHeaders(TickerGet)).Methods("GET")

	router.Path("/realtime").Queries(
		"tickers", "{tickers}",
	).HandlerFunc(wsHandler).Methods("GET")

	router.Path("/login").HandlerFunc(Login).Methods("GET")
	router.Path("/token").HandlerFunc(Token).Methods("GET")
	router.Path("/token").Queries(
		"token", "{token}",
	).HandlerFunc(TokenUpdate).Methods("PUT")

	router.Path("/sendemailotp").HandlerFunc(checkHeaders(SendEmailOTP)).Methods("POST")
	router.Path("/verifyemailotp").HandlerFunc(checkHeaders(VerifyEmailOTP)).Methods("POST")

	router.Path("/").HandlerFunc(HealthCheck).Methods("GET")

	fmt.Println(http.ListenAndServe(":5000", &WithCORS{router}))
}

// WithCORS .
type WithCORS struct {
	r *mux.Router
}

func (s *WithCORS) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "*")
	res.Header().Set("Access-Control-Allow-Headers", "*")

	// Stop here for a Preflighted OPTIONS request.
	if req.Method == "OPTIONS" {
		return
	}

	s.r.ServeHTTP(res, req)
}

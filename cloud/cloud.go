package cloud

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	// postgres driver
	_ "github.com/lib/pq"
)

// User the main structure containing tmp user data
type User struct {
	Useruid       string `json:"user_uid"`
	Useremail     string `json:"user_email"`
	UID           string `json:"uid"`
	Meterid       string
	Utility       string `json:"utility"`
	ServiceTariff string
	Baselineusage float64
	Tariff
	Usage
}

// Tariff the sub User struct containing tariff data for estimating bil;s
type Tariff struct {
	Baseline float64
	Tier2    float64
	Tier3    float64
}

// Usage the sub User struct containing energy and cost data
type Usage struct {
	LastReading   time.Time
	WeekStart     time.Time
	MonthStart    time.Time
	Yesterday     float64
	ThisWeek      float64
	ThisMonth     float64
	CostYesterday float64
	CostThisWeek  float64
	CostThisMonth float64
}

var (
	db     *sql.DB
	dbUser string
	dbPass string
	dbIP   string
	dbDB   string
	dsn    string
	err    error
)

func init() {
	dbUser = os.Getenv("PG_USERNAME")
	dbPass = os.Getenv("PG_PASSWORD")
	dbIP = os.Getenv("PG_EXT_IP")
	dbDB = os.Getenv("PG_DB")
	dsn = "postgres://" + dbUser + ":" + dbPass + "@" + dbIP + ":5432/" + dbDB + "?sslmode=disable"
	// log.Println("pgdb uri: ", dsn)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// log.Println("pgdb connection success")
}

// Read reads from db
func Read(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "connected to db @ %v", dsn)
}

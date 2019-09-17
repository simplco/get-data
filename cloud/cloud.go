package cloud

import (
	"database/sql"
	"fmt"
	"log"
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
	dbUser = os.Getenv("PG_USERNAME")
	dbPass = os.Getenv("PG_PASSWORD")
	dbIN   = os.Getenv("PG_INSTANCE_NAME")
	dsn    = fmt.Sprintf("user=%s password=%s host=/cloudsql/%s", dbUser, dbPass, dbIN)
)

func init() {
	var err error

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("couldnt open db: %v", err)
	}

	// Only allow 1 connection to the database to avoid overloading it.
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
}

// Read reads from db
func Read(w http.ResponseWriter, req *http.Request) {
	// fmt.Fprintf(w, "connected to db @ %v", dsn)

	var result string
	u := User{}

	row, err := db.Query("SELECT * FROM users WHERE email = 'ochoa.erick.d@gmail.com'")
	if err != nil {
		result = "shit, query failed at statement"
		fmt.Fprintln(w, result)
		fmt.Fprintln(w, err)
	}
	for row.Next() {
		err = row.Scan(&u.UID, &u.Meterid, &u.Useremail, &u.Utility, &u.ServiceTariff, &u.LastReading, &u.WeekStart, &u.MonthStart, &u.Yesterday, &u.ThisWeek, &u.ThisMonth, &u.CostYesterday, &u.CostThisWeek, &u.CostThisMonth)
		if err != nil {
			fmt.Println(err)
			result = "shit, query failed at scan"
			fmt.Fprintln(w, result)
		}
	}

	fmt.Fprintf(w, "user: %v", u)
}

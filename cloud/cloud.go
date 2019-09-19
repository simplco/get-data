// TODO: wrap Read in Email fn.

package cloud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	db *sql.DB

	token = os.Getenv("UAPI_TOKEN")

	dbUser = os.Getenv("PG_USERNAME")
	dbPass = os.Getenv("PG_PASSWORD")
	dbIN   = os.Getenv("PG_INSTANCE_NAME")
	dbName = os.Getenv("PG_DB")
	dsn    = fmt.Sprintf("user=%s password=%s host=/cloudsql/%s dbname=%s sslmode=disable", dbUser, dbPass, dbIN, dbName)
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

// Update updates db with latest energy data
func Update(w http.ResponseWriter, req *http.Request) {
	// var result string
	// var err error

	now := time.Now()

	start := now.AddDate(0, 0, -3)
	end := now.AddDate(0, 0, -1)
	u := &User{
		Useremail: "ochoa.erick.d@gmail.com",
	}

	Read(w, req, u)

	fmt.Fprintf(w, "UID: %v\tStarting %v\tEnding: %v\n", u.UID, start.Format("2006-01-02"), end.Format("2006-01-02"))

	readings := getLatestReadingsDay(u.UID, start.Format("2006-01-02"), end.Format("2006-01-02"), token, w)
	calcRecentCosts(readings, u)

	fmt.Fprintf(w, "user: \n%v\n", u)
	prevReading := u.LastReading.AddDate(0, 0, -1)
	fmt.Fprintf(w, "latest reading: %v\t Previous Reading: %v\n", u.LastReading, prevReading)

	_, err := db.Exec("UPDATE users SET (latestts , yescons , wkcons , mocons , yescost , wkcost , mocost) = ($1, $2, $3+$2, $4+$2, $5, $6+$5, $7+$5) WHERE uid = $8 and latestts = $9", u.LastReading, u.Yesterday, u.ThisWeek, u.ThisMonth, u.CostYesterday, u.CostThisWeek, u.CostThisMonth, u.UID, prevReading)
	if err != nil {
		fmt.Fprintln(w, err)
		fmt.Fprintln(w, "damn, insert failed")
	} else {
		fmt.Fprintln(w, "db update success")
	}
}

// Read reads from db
func Read(w http.ResponseWriter, req *http.Request, u *User) {
	// fmt.Fprintf(w, "connected to db @ %v", dsn)

	var result string

	fmt.Fprintln(w, "querying db...")
	row, err := db.Query("SELECT * FROM users WHERE email = $1", u.Useremail)
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

	fmt.Fprintf(w, "user: %v\n", u)
}

func getLatestReadingsDay(uid string, start string, end string, token string, w http.ResponseWriter) []interface{} {
	var url = "https://utilityapi.com/api/v2/intervals?authorizations=" + uid + "&start=" + start + "&end=" + end
	fmt.Fprintf(w, "url: %v\n", url)
	fmt.Fprintf(w, "fetching latest day of intervals for meter %v ...\t", uid)

	intervalRes := makeRequest(url, "GET", token)
	intervals := intervalRes["intervals"].([]interface{})
	list := intervals[0].(map[string]interface{})
	readings := list["readings"].([]interface{})

	fmt.Fprint(w, "ok\n\n")
	return readings
}

func calcRecentCosts(recentInterval []interface{}, u *User) {
	var tariff float64

	if u.ThisMonth >= 339 {
		tariff = 0.39
	}

	tariff = 0.29

	layout := "2006-01-02T15:04:05.999999-07:00"

	// firstReading := readings[len(readings)-1].(map[string]interface{})
	// firstReadingStr, _ := time.Parse(layout, firstReading["end"].(string))
	lastReading := recentInterval[0].(map[string]interface{})
	lastReadingTS, _ := time.Parse(layout, lastReading["end"].(string))
	u.LastReading = lastReadingTS

	var recentTotal float64
	for _, v := range recentInterval[0:24] {
		r := v.(map[string]interface{})
		recentTotal += r["kwh"].(float64)
	}
	recentTotal = math.Floor(recentTotal*100) / 100
	u.Yesterday = recentTotal
	u.CostYesterday = math.Floor(u.Yesterday*tariff*100) / 100
}

func makeRequest(url string, method string, token string) map[string]interface{} {
	// fmt.Println("\t making ", method, " request to:")
	// fmt.Println("\t", url)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", ("Bearer " + token))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, reserr := client.Do(req)
	if reserr != nil {
		fmt.Println("res err")
		log.Fatalln(reserr)
	}
	defer res.Body.Close()

	body, bodyerr := ioutil.ReadAll(res.Body)
	if bodyerr != nil {
		fmt.Println("body err")
		log.Fatalln(bodyerr)
	}

	var data map[string]interface{}
	dataerr := json.Unmarshal(body, &data)
	if dataerr != nil {
		fmt.Println("user err")
		log.Fatalln(dataerr)
	}

	return data
}

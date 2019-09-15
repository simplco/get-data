// TODO: refactor for cloud fn. (create, read, update, delete)
// TODO; refactor for fields in /database/table-script.txt

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	dbUser string
	dbPass string
	dbIP   string
	dbDB   string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("whoops... no .env file")
	}

	var err error

	logFile, logErr := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if logErr != nil {
		log.Fatalf("err opening  file: %v", logErr)
	}

	log.SetOutput(logFile)
	log.Println("LOG START")

	dbUser = os.Getenv("PG_USERNAME")
	dbPass = os.Getenv("PG_PASSWORD")
	dbIP = os.Getenv("PG_EXT_IP")
	dbDB = os.Getenv("PG_DB")
	dsn := "postgres://" + dbUser + ":" + dbPass + "@" + dbIP + ":5432/" + dbDB + "?sslmode=disable"
	log.Println("pgdb uri: ", dsn)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("pgdb connection success")
}

func main() {
	s := time.Now()

	u, _ := getUserData("newbie")
	result := updateDatabase(u)
	fmt.Println(result)

	// data, result := queryDbForUserData("139376")
	data, result := queryDbForUserData(u.UID)
	fmt.Println(result)
	db.Close()

	fmt.Println("UID: ", data.UID, "\tMeter:", data.Meterid, "\tEmail:", data.Useremail, "\tTariff: ", data.ServiceTariff)
	fmt.Println("Latest Reading: ", data.LastReading)
	fmt.Println("Wk TS: ", data.WeekStart)
	fmt.Println("Mo TS: ", data.MonthStart)

	fmt.Println("Yes. Consumption (KWh): ", data.Usage.Yesterday, "\tCost: $", data.Usage.CostYesterday)
	fmt.Println("Wk. Consumption (KWh): ", data.Usage.ThisWeek, "\tCost: $", data.Usage.CostThisWeek)
	fmt.Println("Mo. Consumption (KWh): ", data.Usage.ThisMonth, "\tCost: $", data.Usage.CostThisMonth)

	e := time.Now()
	proctime := e.Sub(s)
	log.Println("took:", proctime)
	log.Println("LOG FINISH")
}

func updateDatabase(d *User) string {
	var result string
	var err error

	_, err = db.Exec("UPDATE users SET ( uid , meter , email , utility , tariff , latestts , wkts , mots , yescons , wkcons , mocons , yescost , wkcost , mocost) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) WHERE uid = $1", d.UID, d.Meterid, d.Useremail, d.Utility, d.ServiceTariff, d.LastReading, d.WeekStart, d.MonthStart, d.Yesterday, d.ThisWeek, d.ThisMonth, d.CostYesterday, d.CostThisWeek, d.CostThisMonth)
	if err != nil {
		fmt.Println(err)
		result = "damn, insert failed"
		return result
	}

	result = "db update success"
	return result
}

func queryDbForUserData(uid string) (User, string) {
	var result string
	u := User{}

	row, err := db.Query("SELECT * FROM users WHERE uid = $1", uid)
	if err != nil {
		fmt.Println(err)
		result = "shit, query failed"
		return u, result
	}
	for row.Next() {
		err = row.Scan(&u.UID, &u.Meterid, &u.Useremail, &u.Utility, &u.ServiceTariff, &u.LastReading, &u.WeekStart, &u.MonthStart, &u.Yesterday, &u.ThisWeek, &u.ThisMonth, &u.CostYesterday, &u.CostThisWeek, &u.CostThisMonth)
		if err != nil {
			fmt.Println(err)
			result = "shit, query failed"
			return u, result
		}
	}

	result = "woo! query sucess"

	return u, result
}

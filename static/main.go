// TODO: refactor for cloud fn. (create, read, update, delete)s

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
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

	simulatedEmail := "ochoa.erick.d@gmail.com"
	result := create(simulatedEmail)
	fmt.Println(result)

	// result = del("ochoa.erick.d@gmail.com")
	// fmt.Println(result)

	data, result := read(simulatedEmail)
	fmt.Println("queried: ", data.UID, "result: ", result)

	db.Close()

	purdyprint(data)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Println("total bytes allocated: ", m.TotalAlloc)

	e := time.Now()
	proctime := e.Sub(s)
	log.Println("took:", proctime)
	log.Println("LOG FINISH")
}

func create(email string) string {
	var result string

	u := getUserData(email)

	if u.UID != "" {
		_, err := db.Exec("insert into users ( uid , meter , email , utility , tariff , latestts , wkts , mots , yescons , wkcons , mocons , yescost , wkcost , mocost) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", u.UID, u.Meterid, u.Useremail, u.Utility, u.ServiceTariff, u.LastReading, u.WeekStart, u.MonthStart, u.Yesterday, u.ThisWeek, u.ThisMonth, u.CostYesterday, u.CostThisWeek, u.CostThisMonth)
		if err != nil {
			fmt.Println(err)
			result = "insert failed"
			return result
		}
		result = "insert successful"
	}

	if u.UID == "" {
		result = "poop, no user found with this uid"
	}
	return result
}

func del(email string) string {
	var result string

	_, err := db.Exec("delete from users where email=$1", email)
	if err != nil {
		log.Println(err)
		result = "poop, deleting user with email:" + email + " failed"
		return result
	}

	result = "deleted user with email: " + email
	return result
}

func update(d *User) string {
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

func read(email string) (User, string) {
	var result string
	u := User{}

	row, err := db.Query("SELECT * FROM users WHERE email = $1", email)
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

func purdyprint(data User) {
	fmt.Println("UID: ", data.UID, "\tMeter:", data.Meterid, "\tEmail:", data.Useremail, "\tTariff: ", data.ServiceTariff)
	fmt.Println("Latest Reading: ", data.LastReading)
	fmt.Println("Wk TS: ", data.WeekStart)
	fmt.Println("Mo TS: ", data.MonthStart)

	fmt.Println("Yes. Consumption (KWh): ", data.Usage.Yesterday, "\tCost: $", data.Usage.CostYesterday)
	fmt.Println("Wk. Consumption (KWh): ", data.Usage.ThisWeek, "\tCost: $", data.Usage.CostThisWeek)
	fmt.Println("Mo. Consumption (KWh): ", data.Usage.ThisMonth, "\tCost: $", data.Usage.CostThisMonth)
}

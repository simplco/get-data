// TODO: add baseline energy fetch/calc
// TODO: calc fn for STANDARD-DR abd TOU-DR1
// TODO: change getLatestIntervalDay() to support updated pricing

// specific parameters
/*
	formuid := user["form_uid"].(string)
	referralcode := user["referrals"].([]interface{})
	email := user["user_email"].(string)
	utility := user["utility"].(string)

	fmt.Println("form uid: ", formuid)
	fmt.Println("referral code: ", referralcode[0])
	fmt.Println("email: ", email)
	fmt.Println("utility: ", utility)
	fmt.Println("user uid: ", useruid)
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

// User the main structure containing tmp user data
type User struct {
	Useruid       string
	Useremail     string
	UID           string
	Meterid       string
	Utility       string
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

func getUserData(usertype string) (*User, string) {
	// start := time.Now()
	u := &User{
		Baselineusage: 339,
		Tariff: Tariff{
			Baseline: 0.29,
			Tier2:    0.39,
			Tier3:    0.55,
		},
	}

	//  use the bearer token
	UtilityBearerToken := os.Getenv("UTILITY_BEARER_TOKEN")

	// returns list of accounts from UtilityAPI
	authdusers := getAuthorizedUsers(UtilityBearerToken)

	// list of authorized users
	auths := authdusers["authorizations"].([]interface{})

	// prune response for necessary info
	user := auths[0].(map[string]interface{})

	u.Useruid = user["user_uid"].(string)
	u.UID = user["uid"].(string)
	u.Utility = user["utility"].(string)
	u.Useremail = user["user_email"].(string)

	// returns user meter metadata
	meterlist := getUserMeter(u.Useruid, UtilityBearerToken)
	// prune response for meter uid
	meter := meterlist["meters"].([]interface{})
	electricmeter := meter[1].(map[string]interface{})
	u.Meterid = electricmeter["uid"].(string)

	base := electricmeter["base"].(map[string]interface{})
	u.ServiceTariff = base["service_tariff"].(string)

	switch usertype {
	case "newbie":
		// returns last 30 days of meter data in hourly intervals
		intervallist := getLatestIntervalsMonth(u.Meterid, "1", UtilityBearerToken)
		calcAllCosts(intervallist, u)
	case "oldie":
		t := time.Now()
		e := t.AddDate(0, 0, -1)
		eStr := e.Format("2006-01-02")

		s := t.AddDate(0, 0, -3)
		sStr := s.Format("2006-01-02")
		fmt.Println("s: ", sStr, "\te: ", eStr)
		intervallist := getLatestReadingsDay(u.UID, sStr, eStr, UtilityBearerToken)
		calcRecentCosts(intervallist, u)
	}

	return u, UtilityBearerToken
}

func getAuthorizedUsers(token string) map[string]interface{} {
	var url = "https://utilityapi.com/api/v2/authorizations?access_token=" + token
	log.Print("fetching authed users...\t")

	userinfo := makeRequest(url, "GET", token)

	log.Print("ok\n")
	return userinfo
}

func getUserMeter(useruid string, token string) map[string]interface{} {
	var url = "https://utilityapi.com/api/v2/meters?users=" + useruid
	log.Print("fetching meter...\t")

	meterinfo := makeRequest(url, "GET", token)

	log.Print("ok\n")
	return meterinfo
}

func getLatestReadingsDay(uid string, start string, end string, token string) []interface{} {
	var url = "https://utilityapi.com/api/v2/intervals?authorizations=" + uid + "&start=" + start + "&end=" + end

	log.Printf("fetching latest day of intervals for meter %v ...\t", uid)

	intervalRes := makeRequest(url, "GET", token)
	intervals := intervalRes["intervals"].([]interface{})
	list := intervals[0].(map[string]interface{})
	readings := list["readings"].([]interface{})

	log.Print("ok\n\n")
	return readings
}

func getLatestIntervalsMonth(meteruid string, limit string, token string) map[string]interface{} {
	var url = "https://utilityapi.com/api/v2/intervals?meters=" + meteruid + "&limit=" + limit + "&order=latest_first"

	log.Printf("fetching intervals for meter %v ...\t", meteruid)

	interval := makeRequest(url, "GET", token)

	log.Print("ok\n")
	return interval
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

func calcAllCosts(pastMonthIntervals map[string]interface{}, u *User) {
	// prune response for meter readings
	intervals := pastMonthIntervals["intervals"].([]interface{})
	list := intervals[0].(map[string]interface{})
	readings := list["readings"].([]interface{})

	var monthlyTotal float64
	var dailyTotal float64
	var weeklyTotal float64

	layout := "2006-01-02T15:04:05.999999-07:00"

	// firstReading := readings[len(readings)-1].(map[string]interface{})
	// firstReadingStr, _ := time.Parse(layout, firstReading["end"].(string))
	lastReading := readings[0].(map[string]interface{})
	lastReadingStr, _ := time.Parse(layout, lastReading["end"].(string))
	dayNum := lastReadingStr.Day()
	weekNum := dayNum % 7

	u.LastReading = lastReadingStr
	u.WeekStart = lastReadingStr.AddDate(0, 0, -weekNum)
	u.MonthStart = lastReadingStr.AddDate(0, 0, -dayNum+1)

	// now := time.Now()

	// numDailyReads := int(now.Sub(firstReadingStr).Hours()/24) - 2

	for i, v := range readings[0 : (dayNum-1)*24+1] {

		r := v.(map[string]interface{})

		if i <= 23 {
			dailyTotal += r["kwh"].(float64)
		}

		if i <= weekNum*24-1 {
			weeklyTotal += r["kwh"].(float64)
		}

		monthlyTotal += r["kwh"].(float64)
	}

	u.Usage.ThisMonth = math.Floor(monthlyTotal*100) / 100
	u.CostThisMonth = math.Floor(u.Usage.ThisMonth*u.Tariff.Baseline*100) / 100

	u.Usage.ThisWeek = math.Floor(weeklyTotal*100) / 100
	u.CostThisWeek = math.Floor(u.Usage.ThisWeek*u.Tariff.Baseline*100) / 100

	u.Usage.Yesterday = math.Floor(dailyTotal*100) / 100
	u.CostYesterday = math.Floor(u.Usage.Yesterday*u.Tariff.Baseline*100) / 100
}

func calcRecentCosts(recentInterval []interface{}, u *User) {
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
	u.CostYesterday = math.Floor(u.Yesterday*u.Baseline*100) / 100
}

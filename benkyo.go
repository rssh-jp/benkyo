package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	TypeConnpass = 1 // connpass
	TypeAtnd     = 2 // atnd
	GoogleApiKey = "AIzaSyCSfHkd-sdQF9MmBHG4rb9j19oLePv3GTU"
)

var (
	JST *time.Location
)

func test() {
	res, err := http.Get("https://connpass.com/api/v1/event/?keyword=golang")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
func httpGet(url string) (ret []byte, err error) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	ret, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

type Data struct {
	providerType   int
	id, title, url string
	start, end     time.Time
}

// ----------------------------------------------------
// connpassの情報取得
// ----------------------------------------------------
// ---------------------------------------
// https://connpass.com/about/api/
// type ResConnpass struct
// type ResConnpassEvent struct
// func getConnpass(from, to time.Time)
// func getConnpassDate(date time.Time, index int)(ret []ResConnpassEvent, err error)
// ---------------------------------------
type ResConnpass struct {
	ResCount    int                `json:"results_returned"`
	AllResCount int                `json:"results_available"`
	SearchNum   int                `json:"results_start"`
	Events      []ResConnpassEvent `json:"events"`
}
type ResConnpassEvent struct {
	EventId int    `json:"event_id"`
	Title   string `json:"title"`
	Start   string `json:"started_at"`
	End     string `json:"ended_at"`
	Url     string `json:"event_url"`
}

func getConnpass(from, to time.Time) (ret []Data, err error) {
	workS := from
	for {
		if to.Sub(workS) <= 0 {
			break
		}
		fmt.Println(workS)

		list, err := getConnpassDate(workS, 1)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, val := range list {
			s, err := time.ParseInLocation(time.RFC3339, val.Start, JST)
			if err != nil {
				fmt.Println(err)
				continue
			}
			e, err := time.ParseInLocation(time.RFC3339, val.End, JST)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ret = append(ret, Data{TypeConnpass, strconv.Itoa(val.EventId), val.Title, val.Url, s, e})
		}

		workS = workS.AddDate(0, 0, 1)
	}
	return
}
func getConnpassDate(date time.Time, index int) (ret []ResConnpassEvent, err error) {
	const reqCount = 10
	var url string
	url = fmt.Sprintf("https://connpass.com/api/v1/event/?ymd=%s&start=%d&count=%d", date.Format("20060102"), index, reqCount)
	res, err := httpGet(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data ResConnpass
	json.Unmarshal(res, &data)
	ret = append(ret, data.Events...)
	if data.AllResCount > data.SearchNum+data.ResCount {
		list, err := getConnpassDate(date, data.SearchNum+data.ResCount)
		if err != nil {
			fmt.Println(err)
			return ret, err
		}
		ret = append(ret, list...)
	}
	return
}

// ----------------------------------------------------
// atndの情報取得
// ----------------------------------------------------
// ---------------------------------------
// http://api.atnd.org/
// type ResConnpass struct
// type ResConnpassEvent struct
// func getConnpass(from, to time.Time)
// func getConnpassDate(date time.Time, index int)(ret []ResConnpassEvent, err error)
// ---------------------------------------
type ResAtnd struct {
	ResCount  int             `json:"results_returned"`
	SearchNum string          `json:"results_start"`
	Events    []ResAtndEvents `json:"events"`
}
type ResAtndEvents struct {
	Event ResAtndEvent `json:"event"`
}
type ResAtndEvent struct {
	EventId int    `json:"event_id"`
	Title   string `json:"title"`
	Start   string `json:"started_at"`
	End     string `json:"ended_at"`
	Url     string `json:"event_url"`
}

func getAtnd(from, to time.Time) (ret []Data, err error) {
	workS := from
	for {
		if to.Sub(workS) <= 0 {
			break
		}
		fmt.Println(workS)

		list, err := getAtndDate(workS, 1)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, val := range list {
			var s, e time.Time
			s, err = time.ParseInLocation(time.RFC3339, val.Start, JST)
			if err != nil {
				fmt.Println(err)
				fmt.Println(val.Title)
				continue
			}
			if val.End == "" {
				e = s.Add(time.Hour * 2)
			} else {
				e, err = time.ParseInLocation(time.RFC3339, val.End, JST)
				if err != nil {
					fmt.Println(err)
					fmt.Println(val.Title)
					continue
				}
			}
			ret = append(ret, Data{TypeAtnd, strconv.Itoa(val.EventId), val.Title, val.Url, s, e})
		}

		workS = workS.AddDate(0, 0, 1)
	}
	return
}
func getAtndDate(date time.Time, index int) (ret []ResAtndEvent, err error) {
	const reqCount = 1
	var url string
	url = fmt.Sprintf("http://api.atnd.org/events/?ymd=%s&start=%d&count=%d&format=json", date.Format("20060102"), index, reqCount)
	res, err := httpGet(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data ResAtnd
	json.Unmarshal(res, &data)
	if data.ResCount > 0 {
		for _, val := range data.Events {
			ret = append(ret, val.Event)
		}
		searchNum, err := strconv.Atoi(data.SearchNum)
		if err != nil {
			fmt.Println(err)
			return ret, err
		}
		list, err := getAtndDate(date, searchNum+data.ResCount)
		if err != nil {
			fmt.Println(err)
			return ret, err
		}
		ret = append(ret, list...)
	}
	return
}
func main() {
	var s, e time.Time
	s = time.Now()

	JST = time.FixedZone("Asia/Tokyo", 9*60*60)
	today := time.Now()
	var startStr, endStr string
	flag.StringVar(&startStr, "s", today.Format("2006-01-02"), "開始日時")
	flag.StringVar(&endStr, "e", today.AddDate(0, 0, 1).Format("2006-01-02"), "終了日時")
	flag.Parse()
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(start, end)

	connpassList, err := getConnpass(start, end)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range connpassList {
		fmt.Printf("%s, %s, %s, %s\n", val.title, val.url, val.start.Format(time.RFC3339), val.end.Format(time.RFC3339))
	}
	fmt.Println(len(connpassList))

	atndList, err := getAtnd(start, end)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range atndList {
		fmt.Printf("%s, %s, %s, %s\n", val.title, val.url, val.start.Format("2006-01-02 15:04:05"), val.end.Format("2006-01-02 15:04:05"))
	}
	fmt.Println(len(atndList))

	e = time.Now()
	fmt.Printf("%f seconds\n", e.Sub(s).Seconds())
}

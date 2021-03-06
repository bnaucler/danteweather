/*

		danteserv.go
		Dante Alighieri giving you the weather report

*/

package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"strconv"
	"net/http"
	"html/template"
	"encoding/json"
	"github.com/jessfraz/weather/forecast"
	"github.com/jessfraz/weather/geocode"
	"github.com/boltdb/bolt"
	"github.com/bnaucler/danteweather/dlib"
)

type Wraw struct { Temp, Hum, Pres, WS, Pint, Pprob float64 }

type Wconv struct { Temp, WS, Pint, Pprob int }

type Winfo struct {
	Loc string
	Temp, Hum, Pres, WS, Pint, Pprob int
}

type Log struct {
	Ltime time.Time
	Lwtr Wraw
	Lloc string
	Lquote string
}

type Scol struct { Col int }

func getwtr(ip string) (cloc string, cwtr Wraw) {

	var (
		geo				geocode.Geocode
		units			string
		ftemp, ctemp	float64
	)

	geo, err := geocode.IPLocate(ip)
	dlib.Cherr(err)

	cloc = fmt.Sprintf("%v, %v", geo.City, geo.CountryCode)

	data := forecast.Request{
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
		Units:     units,
		Exclude:   []string{"hourly", "minutely"},
	}

	fc, err := forecast.Get(fmt.Sprintf("%s/forecast", "https://geocode.jessfraz.com"), data)
	dlib.Cherr(err)

	ftemp = fc.Currently.Temperature
	ctemp = (ftemp - 32) * 5/9

	cwtr = Wraw{Temp: ctemp, Hum: fc.Currently.Humidity,
		Pres: fc.Currently.Pressure, WS: fc.Currently.WindSpeed,
		Pint: fc.Daily.Data[0].PrecipIntensity, Pprob:
		fc.Daily.Data[0].PrecipProbability}

		return
}

func wtrconv(cwtr Wraw) (conv Wconv) {

	conv.Temp = int(cwtr.Temp + 40)
	if conv.Temp < 0 { conv.Temp = 0 }
	if conv.Temp > 99 { conv.Temp = 99 }

	conv.WS = int(cwtr.WS * 3)
	if conv.WS < 0 { conv.WS = 0 }
	if conv.WS > 99 { conv.WS = 99 }

	conv.Pint = int(cwtr.Pint * 2)
	if conv.Pint < 0 { conv.Pint = 0 }
	if conv.Pint > 99 { conv.Pint = 99 }

	conv.Pprob = int(cwtr.Pprob * 10)
	if conv.Pprob < 0 { conv.Pprob = 0 }
	if conv.Pprob > 99 { conv.Pprob = 99 }

	return
}

func verquote(cquote dlib.Quote, cwtrc Wconv) (bool) {

	if cwtrc.Temp < cquote.TempMin || cwtrc.Temp > cquote.TempMax { return false }
	if cwtrc.WS < cquote.WSMin || cwtrc.WS > cquote.WSMax { return false }
	if cwtrc.Pint < cquote.PintMin || cwtrc.Pint > cquote.PintMax { return false }
	if cwtrc.Pprob < cquote.PprobMin || cwtrc.Pprob > cquote.PprobMax { return false }
	return true
}

func caldiff(cquote dlib.Quote, cwtrc Wconv) (diff int) {

	diff = cquote.TempMin + ((cquote.TempMax - cquote.TempMin) / 2)
	diff += cquote.WSMin + ((cquote.WSMax - cquote.WSMin) / 2)
	diff += cquote.PintMin + ((cquote.PintMax - cquote.PintMin) / 2)
	diff += cquote.PprobMin + ((cquote.PprobMax - cquote.PprobMin) / 2)

	diff++ // Average of 1 lost due to divisions. Can be replaced with % if necessary

	return
}

func searchdb (db *bolt.DB, cwtr Wconv, rquote dlib.Quote) (string) {

	tquote := dlib.Quote{}
	mspec := 999
	mdiff := 999

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dlib.Qbuc)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &tquote)
			if tquote.Spec < mspec {
				if verquote(tquote, cwtr) {
					cdiff := caldiff(tquote, cwtr)
					if cdiff < mdiff {
						rquote = tquote
						mspec = tquote.Spec
						mdiff = cdiff
					}
				}
			}
		}
		return nil
	})

	log.Printf("DEBUG: Database entry returned:\n")
	log.Printf("DEBUG: data=%+v\n", rquote)
	log.Printf("DEBUG: mspec=%d\n", mspec)
	log.Printf("DEBUG: mdiff=%d\n", mdiff)

	return rquote.Text
}

func searchlog(db *bolt.DB, k []byte, etime time.Time) (Log, error) {

	etime = etime.Add(-10 * time.Minute)
	clog := Log{}

	err := db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(dlib.Lbuc)
		if buc == nil { return fmt.Errorf("No bucket!") }

		v := buc.Get(k)
		json.Unmarshal(v, &clog)
		return nil
	})

	if clog.Ltime.Before(etime) {
		return Log{}, err
	} else {
		return clog, err
	}
}

func handler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {

	rquote := dlib.Quote{}
	var ip string

	if r.Header.Get("X-Forwarded-For") != "" {
		ip = r.Header["X-Forwarded-For"][0]
		log.Printf("DEBUG: XFW ip: %v\n", ip)
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	log.Printf("Requested path: %v\n", r.URL.Path)
	if(ip == "::1") { ip = "77.249.219.211" } // DEBUG

	// Check database for hit on IP Within time range
	now := time.Now()
	clog, err := searchlog(db, []byte(ip), now)

	if len(clog.Lquote) == 0 || err != nil {
		cloc, rwtr := getwtr(ip)
		cwtr := wtrconv(rwtr)
		clog = Log{now, rwtr, cloc, searchdb(db, cwtr, rquote)}
		log.Printf("DEBUG: clog: %+v", clog)
		v, err := json.Marshal(clog)
		dlib.Cherr(err)
		err = dlib.Wrdb(db, []byte(ip), v, dlib.Lbuc)
		dlib.Cherr(err)
		log.Printf("Serving %v (%v) with new data\n", string(ip), clog.Lloc)
	} else {
		log.Printf("Serving %v (%v) from database\n", string(ip), clog.Lloc)
	}

	// Yes there is a smarter way to do this
	if r.URL.Path == "/default.css" {
		chr, _ := strconv.Atoi(now.Format("15"))
		cmn, _ := strconv.Atoi(now.Format("04"))
		t, _ := template.ParseFiles("html/default.css")
		w.Header().Add("Content-Type", "text/css")
		t.Execute(w, Scol{Col: ((chr * 60) + cmn) / 5})
	} else if r.URL.Path == "/raw" {
		t, _ := template.ParseFiles("html/raw.md")
		t.Execute(w, clog)
	} else if r.URL.Path == "/info" {
		winf := Winfo{Loc: clog.Lloc, Temp: fltoint(clog.Lwtr.Temp),
			Hum: fltoint((clog.Lwtr.Hum * 100)), Pres: fltoint(clog.Lwtr.Pres),
			WS: fltoint(clog.Lwtr.WS), Pint: fltoint(clog.Lwtr.Pint),
			Pprob: fltoint(clog.Lwtr.Pprob * 100)}
		t, _ := template.ParseFiles("html/info.html")
		t.Execute(w, winf)
	} else if r.URL.Path == "/favicon.ico" {
		// Do nothing
	} else {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, clog)
	}
}

func fltoint (input float64) int {

	tstr := fmt.Sprintf("%.0f", input)
	output, err := strconv.Atoi(tstr)

	if err == nil { return output
	} else { return 0 }
}

func main() {

	db, err := bolt.Open(dlib.DBname, 0640, nil)
	dlib.Cherr(err)
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, db)
	})

	err = http.ListenAndServe(":8959", nil)
	dlib.Cherr(err)
}

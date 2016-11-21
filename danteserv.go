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
	"net/http"
	"encoding/json"
	"github.com/jessfraz/weather/forecast"
	"github.com/jessfraz/weather/geocode"
	"github.com/boltdb/bolt"
	"github.com/bnaucler/danteweather/dlib"
)

var (
	dbname = string("./dante.db")
	qbuc = []byte("quotes")
	lbuc = []byte("visitors")
)

type Wraw struct { temp, hum, pres, ws, pint, pprob float64 }

type Wconv struct { Temp, WS, Pint, Pprob int }

type Log struct {
	Ltime time.Time
	Lwtr Wraw
	Lloc string
	Lquote string
}

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

	cwtr = Wraw{temp: ctemp, hum: fc.Currently.Humidity,
		pres: fc.Currently.Pressure, ws: fc.Currently.WindSpeed,
		pint: fc.Daily.Data[0].PrecipIntensity, pprob:
		fc.Daily.Data[0].PrecipProbability}

		return
}

func wtrconv(cwtr Wraw) (conv Wconv) {

	conv.Temp = int(cwtr.temp + 40)
	if conv.Temp < 0 { conv.Temp = 0 }
	if conv.Temp > 99 { conv.Temp = 99 }

	conv.WS = int(cwtr.ws * 3)
	if conv.WS < 0 { conv.WS = 0 }
	if conv.WS > 99 { conv.WS = 99 }

	conv.Pint = int(cwtr.pint * 2)
	if conv.Pint < 0 { conv.Pint = 0 }
	if conv.Pint > 99 { conv.Pint = 99 }

	conv.Pprob = int(cwtr.pprob * 10)
	if conv.Pprob < 0 { conv.Pprob = 0 }
	if conv.Pprob > 99 { conv.Pprob = 99 }

	log.Printf("DEBUG: converted values: %+v", conv)

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

	diff += 1 // Average of 1 lost due to divisions. Can be replaced with % if necessary

	return
}

func searchdb (db *bolt.DB, cwtr Wconv, rquote dlib.Quote) (string) {

	tquote := dlib.Quote{}
	mspec := 999
	mdiff := 999

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(qbuc)
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
						log.Printf("DEBUG: mspec=%d\n", mspec)
						log.Printf("DEBUG: mdiff=%d\n", mdiff)
					}
				}
			}
		}
		return nil
	})
	return rquote.Text
}

func searchlog(db *bolt.DB, k []byte, etime time.Time) (Log, error) {

	etime = etime.Add(-10 * time.Minute)
	clog := Log{}

	err := db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(lbuc)
		if buc == nil { return fmt.Errorf("No bucket!") }

		v := buc.Get(k)
		json.Unmarshal(v, &clog)
		return nil
	})

	log.Printf("DEBUG: ctime: %v", etime)
	log.Printf("DEBUG: clog.Ltime: %v", clog.Ltime)

	if clog.Ltime.Before(etime) {
		return Log{}, err
	} else {
		return clog, err
	}
}

func handler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {

	rquote := dlib.Quote{}

	var (
		cloc string
		quote string
		rwtr Wraw
		cwtr Wconv
	)

	log.Printf("Requested path: %v\n", r.URL.Path)
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if(ip == "10.0.0.20") { ip = "77.249.219.212" }

	// Check database for hit on IP Within time range
	now := time.Now()
	clog, err := searchlog(db, []byte(ip), now)

	if len(clog.Lquote) == 0 || err != nil {
		cloc, rwtr = getwtr(ip)
		cwtr = wtrconv(rwtr)
		quote = searchdb(db, cwtr, rquote)
		wlog := Log{now, rwtr, cloc, quote}
		v, err:= json.Marshal(wlog)
		dlib.Cherr(err)
		err = dlib.Wrdb(db, []byte(ip), v, lbuc)
		dlib.Cherr(err)
		log.Printf("DEBUG: Serving %v with new data\n", string(ip))
	} else {
		cloc = clog.Lloc
		rwtr = clog.Lwtr
		quote = clog.Lquote
		log.Printf("DEBUG: Serving %v with data from database\n", string(ip))
	}

	fmt.Fprintf(w, "Weather for your location according to Dante:\n%s\n", 
	quote)

	fmt.Fprintf(w, "In other words: weather for %v\n", cloc)
	fmt.Fprintf(w, "Temperature: %.2f\n", rwtr.temp)
	fmt.Fprintf(w, "Humidity: %.2f\n", rwtr.hum)
	fmt.Fprintf(w, "Pressure: %.2f\n", rwtr.pres)
	fmt.Fprintf(w, "Wind speed: %.2f\n", rwtr.ws)
	fmt.Fprintf(w, "Precipitation intensity: %.5f\n", rwtr.pint)
	fmt.Fprintf(w, "Precipitation probability: %.2f\n", rwtr.pprob)
}

func main() {

	db, err := bolt.Open(dbname, 0640, nil)
	dlib.Cherr(err)
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, db)
	})

	err = http.ListenAndServe(":8959", nil)
	dlib.Cherr(err)
}

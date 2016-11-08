/*

		danteserv.go
		Dante Alighieri giving you the weather report

*/

package main

import (
	"fmt"
	"log"
	"net"
	// "strings"
	"net/http"
	"encoding/json"
	"github.com/jessfraz/weather/forecast"
	"github.com/jessfraz/weather/geocode"
	"github.com/boltdb/bolt"
)

var	cbuc = []byte("bucket")

func cherr(e error) {
	if e != nil { panic(e) }
}

type Wraw struct { temp, hum, pres, ws, pint, pprob float64 }

type Wconv struct { Temp, WS, Pint, Pprob int }

type quote struct {
	TempMin, TempMax int
	WSMin, WSMax int
	PintMin, PintMax int
	PprobMin, PprobMax int
	Spec int
	Text string
}

func getwtr(ip string) (cloc string, cwtr Wraw) {

	var (
		geo				geocode.Geocode
		units			string
		ftemp, ctemp	float64
	)

	geo, err := geocode.IPLocate(ip)
	cherr(err)

	cloc = fmt.Sprintf("%v, %v", geo.City, geo.CountryCode)

	data := forecast.Request{
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
		Units:     units,
		Exclude:   []string{"hourly", "minutely"},
	}

	fc, err := forecast.Get(fmt.Sprintf("%s/forecast", "https://geocode.jessfraz.com"), data)
	cherr(err)

	ftemp = fc.Currently.Temperature
	ctemp = (ftemp - 32) * 5/9

	cwtr = Wraw{temp: ctemp, hum: fc.Currently.Humidity,
		pres: fc.Currently.Pressure, ws: fc.Currently.WindSpeed,
		pint: fc.Daily.Data[0].PrecipIntensity, pprob:
		fc.Daily.Data[0].PrecipProbability}

		return
}

func wtrconv(cwtr Wraw) (conv Wconv) {

	log.Printf("DEBUG: opening wtrconv")

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

	log.Printf("DEBUG: closing wtrconv")
	log.Printf("DEBUG: converted values: %+v", conv)

	return
}

func verquote(cquote quote, cwtrc Wconv) (bool) {

	if cwtrc.Temp < cquote.TempMin || cwtrc.Temp > cquote.TempMax { return false }
	if cwtrc.WS < cquote.WSMin || cwtrc.WS > cquote.WSMax { return false }
	if cwtrc.Pint < cquote.PintMin || cwtrc.Pint > cquote.PintMax { return false }
	if cwtrc.Pprob < cquote.PprobMin || cwtrc.Pprob > cquote.PprobMax { return false }
	return true
}

func caldiff(cquote quote, cwtrc Wconv) (diff int) {

	diff = cquote.TempMin + ((cquote.TempMax - cquote.TempMin) / 2)
	diff += cquote.WSMin + ((cquote.WSMax - cquote.WSMin) / 2)
	diff += cquote.PintMin + ((cquote.PintMax - cquote.PintMin) / 2)
	diff += cquote.PprobMin + ((cquote.PprobMax - cquote.PprobMin) / 2)

	diff += 1 // Average of 1 lost due to divisions. Can be replaced with % if necessary

	return
}

func searchdb (db *bolt.DB, cwtr Wraw, rquote quote) (string) {

	curr := wtrconv(cwtr)
	tquote := quote{}
	mspec := 999
	mdiff := 999

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cbuc)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &tquote)
			if tquote.Spec < mspec {
				if verquote(tquote, curr) {
					cdiff := caldiff(tquote, curr)
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

func handler(w http.ResponseWriter, r *http.Request, db *bolt.DB) {

	rquote := quote{}

	log.Printf("Requested path: %v\n", r.URL.Path)
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if(ip == "10.0.0.20") { ip = "77.249.219.212" }

	cloc, cwtr := getwtr(ip)

	log.Printf("Remote host: %v\t%v\n", ip, cloc)

	quote := searchdb(db, cwtr, rquote)
	fmt.Fprintf(w, "Weather for your location according to Dante:\n%s\n", quote)

	fmt.Fprintf(w, "In other words: weather for %v\n", cloc)
	fmt.Fprintf(w, "Temperature: %.2f\n", cwtr.temp)
	fmt.Fprintf(w, "Humidity: %.2f\n", cwtr.hum)
	fmt.Fprintf(w, "Pressure: %.2f\n", cwtr.pres)
	fmt.Fprintf(w, "Wind speed: %.2f\n", cwtr.ws)
	fmt.Fprintf(w, "Precipitation intensity: %.5f\n", cwtr.pint)
	fmt.Fprintf(w, "Precipitation probability: %.2f\n", cwtr.pprob)
}

func main() {

	dbname := "./dante.db"

	db, err := bolt.Open(dbname, 0640, nil)
	cherr(err)
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, db)
	})

	err = http.ListenAndServe(":8959", nil)
	cherr(err)
}

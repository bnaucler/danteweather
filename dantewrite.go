/*

		dantewrite.go
		Helper program for danteweather. Writes an entry to the database

*/

package main

import (
	"fmt"
	"os"
	"bufio"
	"strconv"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/bnaucler/danteweather/dlib"
)

func rval(prompt string, minval, maxval int) (chint int) {

	var tmp string
	fmt.Printf("%v: ", prompt)
	fmt.Scanln(&tmp)

	chint, err := strconv.Atoi(tmp)
	if err != nil { panic("Not a number") 
	} else if chint < minval || chint > maxval {
		resp := fmt.Sprintf("Number not in %d-%d range", minval, maxval)
		panic(resp)
	}

	dlib.Cherr(err)
	return
}

func rtext(prompt string, eof string) ([]byte) {

	var sbuf string
	var bbuf string

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println(prompt)

	for {
		sbuf = ""
		scanner.Scan()
		sbuf = scanner.Text()
		if sbuf != eof {
			if len(bbuf) == 0 {
				bbuf = sbuf
			} else {
				bbuf = fmt.Sprintf("%v\n%v", bbuf, sbuf)
			}
		} else { break }
	}
	return []byte(bbuf)
}

func main() {

	qbuc := []byte("quotes")
	eof := "EOF"
	cquote := dlib.Quote{}
	rquote := dlib.Quote{}

	db, err := bolt.Open(dlib.DBname, 0640, nil)
	dlib.Cherr(err)
	defer db.Close()

	fmt.Print("Enter key: ")
	var tmp string
	fmt.Scanln(&tmp)
	k := []byte(tmp)

	fmt.Println("Enter values 0-99")

	cquote.TempMin = rval("Temperature min", 0, 99)
	cquote.TempMax = rval("Temperature max", 0, 99)
	if cquote.TempMin > cquote.TempMax { panic("Minimum larger than maximum!") }
	cquote.WSMin = rval("Wind speed min", 0, 99)
	cquote.WSMax = rval("Wind speed max", 0, 99)
	if cquote.WSMin > cquote.WSMax { panic("Minimum larger than maximum!") }
	cquote.PintMin = rval("Precipitation intensity min", 0, 99)
	cquote.PintMax = rval("Precipitation intensity max", 0, 99)
	if cquote.PintMin > cquote.PintMax { panic("Minimum larger than maximum!") }
	cquote.PprobMin = rval("Precipitation probability min", 0, 99)
	cquote.PprobMax = rval("Precipitation probability max", 0, 99)
	if cquote.PprobMin > cquote.PprobMax { panic("Minimum larger than maximum!") }

	cquote.Spec = (cquote.TempMax - cquote.TempMin) +
		(cquote.WSMax - cquote.WSMin) +
		(cquote.PintMax - cquote.PintMin) +
		(cquote.PprobMax - cquote.PprobMin)

	prompt := fmt.Sprintf("Enter text - end with %s", eof)
	cquote.Text = string(rtext(prompt, eof))

	v, err := json.Marshal(cquote)
	dlib.Cherr(err)

	err = dlib.Wrdb(db, k, v, qbuc)
	dlib.Cherr(err)

	val, err := dlib.Rdb(db, k, qbuc)
	dlib.Cherr(err)

	json.Unmarshal(val, &rquote)

	fmt.Printf("Complete entry as read from %v:\n%v:\t%+v\n", dlib.DBname, string(k), rquote)
}

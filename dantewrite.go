/*

		dantewrite.go
		Helper program for danteweather. Writes an entry to the database

*/

package main

import (
	"fmt"
	"os"
	"bufio"
	"unicode"
	"strconv"
	"encoding/json"
	"github.com/boltdb/bolt"
)

type quote struct {
	TempMin, TempMax int
	WSMin, WSMax int
	PintMin, PintMax int
	PprobMin, PprobMax int
	Spec int
	Text string
}

func cherr(e error) {
	if e != nil { panic(e) }
}

func wrdb(db *bolt.DB, k, v, cbuc []byte) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		buc, err := tx.CreateBucketIfNotExists(cbuc)
		if err != nil { return err }

		err = buc.Put(k, v)
		if err != nil { return err }

		return nil
	})
	return
}

func rdb(db *bolt.DB, k, cbuc []byte) (v []byte, err error) {

	err = db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(cbuc)
		if buc == nil { return fmt.Errorf("No bucket!") }

		v = buc.Get(k)
		return nil
	})
	return
}

func rval(prompt string, minval, maxval int) (chint int) {

	var tmp string
	fmt.Printf("%v: ", prompt)
	fmt.Scanln(&tmp)
	for a := 0; a < len(tmp); a++ {
		r := rune(tmp[a])
		if !unicode.IsDigit(r) {
			panic("Not a number")
		}
	}

	chint, err := strconv.Atoi(tmp)
	if chint < minval || chint > maxval {
		panic("Number not in 0-99 range")
	}

	cherr(err)
	return
}

func rtext(prompt string) ([]byte) {

	var sbuf string
	var bbuf string

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println(prompt)

	for {
		sbuf = ""
		scanner.Scan()
		sbuf = scanner.Text()
		if sbuf != "EOF" {
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

	cbuc := []byte("quotes")
	dbname := "./dante.db"
	cquote := quote{}
	rquote := quote{}

	db, err := bolt.Open(dbname, 0640, nil)
	cherr(err)
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

	cquote.Text = string(rtext("Enter text - end with EOF"))

	v, err := json.Marshal(cquote)
	cherr(err)

	err = wrdb(db, k, v, cbuc)
	cherr(err)

	val, err := rdb(db, k, cbuc)
	cherr(err)

	json.Unmarshal(val, &rquote)

	fmt.Printf("Complete entry as read back from %v:\n%v:\t%+v\n", dbname, string(k), rquote)
}

/* 

		dantedump.go
		Helper program for danteweather. Dumps database to console for troubleshooting

*/

package main

import (
	"fmt"
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

func dbdump (db *bolt.DB, cbuc []byte, rquote quote) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(cbuc)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &rquote)
			fmt.Printf("%s=%+v\n", k, rquote)
		}
		return nil
	})
}

func main() {

	cbuc := []byte("quotes")
	dbname := "./dante.db"
	rquote := quote{}

	db, err := bolt.Open(dbname, 0640, nil)
	cherr(err)
	defer db.Close()

	dbdump(db, cbuc, rquote)
}

/* 

		dantedump.go
		Helper program for danteweather. Dumps database to console for troubleshooting

*/

package main

import (
	"fmt"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/bnaucler/danteweather/dlib"
)

func dbdump (db *bolt.DB, qbuc []byte, rquote *dlib.Quote) {

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(qbuc)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &rquote)
			fmt.Printf("%s=%+v\n", k, rquote)
		}
		return nil
	})
}

func main() {

	qbuc := []byte("quotes")
	rquote := dlib.Quote{}

	db, err := bolt.Open(dlib.DBname, 0640, nil)
	dlib.Cherr(err)
	defer db.Close()

	dbdump(db, qbuc, &rquote)
}

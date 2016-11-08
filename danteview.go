/*

		danteview.go
		Helper program for danteweather. Displays one db value on demand

*/

package main

import (
	"fmt"
	"os"
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

func rdb(db *bolt.DB, k []byte, rquote quote, cbuc []byte) (quote, error) {

	err := db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(cbuc)
		if buc == nil { return fmt.Errorf("No bucket!") }

		v := buc.Get(k)
		json.Unmarshal(v, &rquote)
		return nil
	})
	return rquote, err
}

func main() {

	argc := len(os.Args)

	if argc != 2 {
		emsg := fmt.Sprintf("Usage: %v <key>", os.Args[0])
		panic(emsg)
	}

	cbuc := []byte("bucket")
	dbname := "./dante.db"
	rquote := quote{}

	db, err := bolt.Open(dbname, 0640, nil)
	cherr(err)
	defer db.Close()

	rquote, err = rdb(db, []byte(os.Args[1]), rquote, cbuc)
	cherr(err)

	fmt.Printf("%v:\n%+v\n", os.Args[1], rquote)
}

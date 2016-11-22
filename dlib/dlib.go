package dlib

import (
	"fmt"
	"github.com/boltdb/bolt"
)

var (
	DBname = string("./dante.db")
	Qbuc = []byte("quotes")
	Lbuc = []byte("visitors")
)

type Quote struct {
	TempMin, TempMax int
	WSMin, WSMax int
	PintMin, PintMax int
	PprobMin, PprobMax int
	Spec int
	Text string
}

func Cherr(e error) {
	if e != nil { panic(e) }
}

func Wrdb(db *bolt.DB, k, v, cbuc []byte) (err error) {

	err = db.Update(func(tx *bolt.Tx) error {
		buc, err := tx.CreateBucketIfNotExists(cbuc)
		if err != nil { return err }

		err = buc.Put(k, v)
		if err != nil { return err }

		return nil
	})
	return
}

func Rdb(db *bolt.DB, k, cbuc []byte) (v []byte, err error) {

	err = db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(cbuc)
		if buc == nil { return fmt.Errorf("No bucket!") }

		v = buc.Get(k)
		return nil
	})
	return
}


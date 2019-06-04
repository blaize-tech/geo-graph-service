package item

import (
	"fmt"
	"log"
	"time"

	"geo-graph-service/api/models/item/db"

	"gopkg.in/mgo.v2/bson"
)

var equi uint32

//Trustline object
type Trustline struct {
	Source      string    `json:"source" bson:"nodeSource"`
	Destination string    `json:"destination" bson:"nodeDestination"`
	Equivalent  uint32    `json:"equivalent" bson:"equivalent" `
	State       string    `json:"state" bson:"state"`
	Time        time.Time `bson:"date"`
}

type TrustlineAPI struct {
	Contractor  string    `json:"contractor"`
	Equivalent  uint32    `json:"equivalent_id"`
	SetDateTime time.Time `json:"setdatetime"`
}

// PostTrustline saves trustline (form data) into the database.
func PostTrustline(trustline *Trustline) error {
	_, err := FindTrustline(trustline.Destination, trustline.Source)
	if err != nil {
		return fmt.Errorf("'Trustline is already exists in db'")
	}
	trustline.Equivalent = equi
	trustline.State = "on"
	if err := db.SaveItem(trustline, "trustline"); err != nil {
		return err
	}
	if err := db.SaveItem(trustline, "trustline_history"); err != nil {
		return err
	}

	return nil
}

//DeleteTrustline removes trustline from db
func DeleteTrustline(src string, dst string) (string, string, error) {
	toDelete, err := FindTrustline(src, dst)
	var sr, dr string
	if err != nil {
		err = removeTrustline(src, dst, "trustline")
		sr = src
		dr = dst
		if err != nil {
			err = removeTrustline(dst, src, "trustline")
			sr = dst
			dr = src
			if err != nil {
				return "", "", err
			}
		}
	}
	toDelete.State = "off"
	toDelete.Time = time.Now()
	if err := db.SaveItem(toDelete, "trustline_history"); err != nil {
		return "", "", err
	}

	return sr, dr, nil
}

//RemoveAllTrustlines removes all trustline from db
func RemoveAllTrustlines() error {
	trustlines, err := GetAllTrustlines()
	if err != nil {
		return fmt.Errorf("Cant load trustlines from db: %v", err)
	}
	for i := range trustlines {
		err := removeTrustline(trustlines[i].Source, trustlines[i].Destination, "trustline")
		if err != nil {
			log.Printf("range removing trustlines error occured: %v", err)
		}
	}
	return nil
}

// GetAllTrustlines returns all trustline items from the table of database.
func GetAllTrustlines() ([]Trustline, error) {
	res := []Trustline{}
	if err := db.GetCollection("trustline").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

//FindTrustline serach for trustline, if db exists one returns trustline object
func FindTrustline(hashSource string, hashDestination string) (*Trustline, error) {
	res := Trustline{}

	if err := db.GetCollection("trustline").Find(bson.M{"nodeSource": hashSource, "nodeDestination": hashDestination}).One(&res); err == nil {
		return &res, fmt.Errorf("Trustline is already exists")
	}

	if err := db.GetCollection("trustline").Find(bson.M{"nodeSource": hashDestination, "nodeDestination": hashSource}).One(&res); err == nil {
		return &res, fmt.Errorf("Trustline is already exists")
	}

	return &res, nil
}

// remove deletes an item from the table of database
func removeTrustline(hash string, hashWith string, tableName string) error {
	return db.GetCollection(tableName).Remove(bson.M{"nodeSource": hash, "nodeDestination": hashWith})
}

func getTrustlinesByDestination(hash string) ([]Trustline, error) {
	res := []Trustline{}
	if err := db.GetCollection("trustline").Find(bson.M{"nodeHashWith": hash}).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func getTrustlinesBySource(hash string) ([]Trustline, error) {
	res := []Trustline{}
	if err := db.GetCollection("trustline").Find(bson.M{"nodeHash": hash}).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

// getTrustline returns a single item from the database.
func getTrustline(hash string) (*Trustline, error) {
	res := Trustline{}

	if err := db.GetCollection("trustline").Find(bson.M{"nodeHash": hash}).One(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func TrustlineRepacker(old Trustline) (newer TrustlineAPI) {
	newer.Contractor = old.Destination
	newer.Equivalent = old.Equivalent
	newer.SetDateTime = old.Time
	return
}

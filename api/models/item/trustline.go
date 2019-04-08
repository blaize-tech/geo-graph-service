package item

import (
	"fmt"
	"log"
	"time"

	"github.com/GeoServer/project/api/models/item/db"
	"gopkg.in/mgo.v2/bson"
)

var equi uint32

//Trustline object
type Trustline struct {
	Source      string    `json:"source" bson:"nodeSource"`
	Destination string    `json:"destination" bson:"nodeDestination"`
	Equivalent  uint32    `json:"equivalent" bson:"equivalent" `
	Time        time.Time `bson:"time"`
}

// PostTrustline saves trustline (form data) into the database.
func PostTrustline(trustline *Trustline) error {
	// if trustline.Source == trustline.Destination {
	// 	return fmt.Errorf("destination and source are the same")
	// }

	_, err := FindTrustline(trustline.Destination, trustline.Source)
	if err != nil {
		return fmt.Errorf("'Trustline is already exists in db'")
	}
	trustline.Equivalent = equi
	if err := db.SaveItem(trustline, "trustline"); err != nil {
		return err
	}
	return nil

}

//DeleteTrustline removes trustline from db
func DeleteTrustline(src string, dst string) error {
	_, err := FindTrustline(src, dst)
	if err != nil {
		if err = removeTrustline(src, dst, "trustline"); err != nil {
			if err = removeTrustline(dst, src, "trustline"); err != nil {
				return err
			}
		}
	}
	return nil
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
		return nil, fmt.Errorf("Trustline is already exists")
	}

	if err := db.GetCollection("trustline").Find(bson.M{"nodeSource": hashDestination, "nodeDestination": hashSource}).One(&res); err == nil {
		return nil, fmt.Errorf("Trustline is already exists")
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

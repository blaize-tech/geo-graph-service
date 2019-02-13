package main

import (
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Trustline struct {
	Source       	string  	`json:"nodeHashFrom" bson:"nodeHash"`
	Destination 	string  	`json:"nodeHashTo" bson:"nodeHashWith"`
	Op     			bool    	`json:"op" bson:"op"`
	Time      		time.Time	`bson:"time"`
}

type Payment struct {
	Source       	string  	`json:"fromNodeHash" bson:"nodeHash"`
	Destination 	string  	`json:"toNodeHash" bson:"nodeHashWith"`
	Time      		time.Time	`bson:"time"`
	Paths     		[]string   	`json:"paths" bson:"pathHashs"`
}

var db *mgo.Database

func init() {
	session, err := mgo.Dial("localhost/api_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db = session.DB("api_db")
}
// getCollection return collection from database
// trustline payment
func getCollection(tableName string) *mgo.Collection {
	return db.C(tableName)
}

// getAll returns all items from the table of database.
func getAllTrustlines() ([]Trustline, error) {
	res := []Trustline{}
	if err := getCollection("trustline").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func getAllPayments() ([]Payment, error) {
	res := []Payment{}
	if err := getCollection("payment").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

// getTrustline returns a single item from the database.
func getTrustline(hash string) (*Trustline, error) {
	res := Trustline{}

	if err := getCollection("trustline").Find(bson.M{"nodeHash": hash}).One(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
// getTrustline returns a single item from the database.
func getPayment(hash string) (*Payment, error) {
	res := Payment{}
	if err := getCollection("payment").Find(bson.M{"nodeHash": hash}).One(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
/**
	@dev
 */
func updateTrustline(t *Trustline) error {
	colQuerier := bson.M{"nodeHash": t.Source}
	err := getCollection("trustline").Update(colQuerier, &t)
	if err != nil {
		return err
	}
	return nil
}

// save inserts an item to the database.
func saveItem(item interface{}, tableName string) error {
	return getCollection(tableName).Insert(item)
}

// remove deletes an item from the table of database
func removeItem(hash string, tableName string) error {
	return getCollection(tableName).Remove(bson.M{"nodeHash": hash})
}

func clearAll() error {

	rsTrustlines, err := getAllTrustlines()
	if err != nil {
		log.Println("Failed to load database items:", err)
		return err
	}
	for i:=0;i<len(rsTrustlines);i++ {
		removeItem(rsTrustlines[i].Source, "trustline")
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println("Failed to load database items:", err)
		return err
	}
	for i:=0;i<len(rsPayments);i++ {
		removeItem(rsPayments[i].Source, "payment")
	}
	return nil
}

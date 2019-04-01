package main

import (
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Hash string `json:"hash" bson:"hash"`
}

type Trustline struct {
	//Id				string		`bson:"id"`
	Source      string `json:"source" bson:"nodeSource"`
	Destination string `json:"destination" bson:"nodeDestination"`
	//	Delete			bool
	Equivalent uint32    `json:"equivalent" bson:"equivalent" `
	Time       time.Time `bson:"time"`
}

// mock for Trustline.Equivalent data
var equi uint32 = 0

type Payment struct {
	Source      string `json:"source" bson:"nodeHash"`
	Destination string `json:"destination" bson:"nodeHashWith"`
	// Source      string     `json:"nodeHashFrom" bson:"nodeHash"`
	// Destination string     `json:"nodeHashTo" bson:"nodeHashWith"`
	Time  time.Time  `bson:"time"`
	Paths [][]string `json:"paths" bson:"pathHashs"`
}

var db *mgo.Database

func init() {
	session, err := mgo.Dial("172.17.0.1:25017/api_db")
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

func getAllNodes() ([]Node, error) {
	res := []Node{}
	if err := getCollection("nodes").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
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

func getTrustlinesByDestination(hash string) ([]Trustline, error) {
	res := []Trustline{}
	if err := getCollection("trustline").Find(bson.M{"nodeHashWith": hash}).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func getTrustlinesBySource(hash string) ([]Trustline, error) {
	res := []Trustline{}
	if err := getCollection("trustline").Find(bson.M{"nodeHash": hash}).All(&res); err != nil {
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
func removeItem(hash string, hashWith string, tableName string) error {
	return getCollection(tableName).Remove(bson.M{"nodeSource": hash, "nodeDestination": hashWith})
}

func removeNode(hash string, tableName string) error {
	return getCollection(tableName).Remove(bson.M{"hash": hash})
}

func clearAll() error {

	rsTrustlines, err := getAllTrustlines()
	log.Println(len(rsTrustlines))
	if err != nil {
		log.Println("Failed to load database trustlines:", err)
		return err
	}
	for i := 0; i < len(rsTrustlines); i++ {
		log.Println(removeItem(rsTrustlines[i].Source, rsTrustlines[i].Destination, "trustline"))
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println("Failed to load database payments:", err)
		return err
	}
	for i := 0; i < len(rsPayments); i++ {
		log.Println(removeItem(rsPayments[i].Source, rsTrustlines[i].Destination, "payment"))
	}

	rsNodes, err := getAllNodes()
	if err != nil {
		log.Println("Failed to load database nodes:", err)
		return err
	}
	for i := 0; i < len(rsNodes); i++ {
		log.Println(removeNode(rsNodes[i].Hash, "nodes"))
	}

	return nil
}

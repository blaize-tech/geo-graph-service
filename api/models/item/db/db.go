package db

import (
	"log"

	"gopkg.in/mgo.v2"
)

var dB *mgo.Database

//Initialize connection
func init() {
	session, err := mgo.Dial("localhost:27017/api_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	dB = session.DB("api_db")
}

// GetCollection return collection from database
func GetCollection(tableName string) *mgo.Collection {
	return dB.C(tableName)
}

//SaveItem inserts an item to the database.
func SaveItem(item interface{}, tableName string) error {
	return GetCollection(tableName).Insert(item)
}

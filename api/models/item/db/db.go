package db

import (
	"log"

	"gopkg.in/mgo.v2"
)

var db *mgo.Database

//Initialize connection
func InitDB() {
	session, err := mgo.Dial("172.17.0.1:25017/api_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db = session.DB("api_db")
}

// getCollection return collection from database
// trustline payment
func GetCollection(tableName string) *mgo.Collection {
	return db.C(tableName)
}

/**
@dev
*/

// save inserts an item to the database.
func SaveItem(item interface{}, tableName string) error {
	return GetCollection(tableName).Insert(item)
}

/*func clearAll() error {

	rsTrustlines, err := getAllTrustlines()
	log.Println(len(rsTrustlines))
	if err != nil {
		log.Println("Failed to load database trustlines:", err)
		return err
	}
	for i := 0; i < len(rsTrustlines); i++ {
		log.Println(removeTrustline(rsTrustlines[i].Source, rsTrustlines[i].Destination, "trustline"))
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println("Failed to load database payments:", err)
		return err
	}
	for i := 0; i < len(rsPayments); i++ {
		log.Println(removeTrustline(rsPayments[i].Source, rsTrustlines[i].Destination, "payment"))
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
}*/

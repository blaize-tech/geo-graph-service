package item

import (
	"time"

	"github.com/geo-graph-service/api/models/item/db"
	"gopkg.in/mgo.v2/bson"
)

type Payment struct {
	Source      string     `json:"source" bson:"nodeHash"`
	Destination string     `json:"destination" bson:"nodeHashWith"`
	Time        time.Time  `bson:"time"`
	Paths       [][]string `json:"paths" bson:"pathHashs"`
}

func GetAllPayments() ([]Payment, error) {
	res := []Payment{}
	if err := db.GetCollection("payment").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

// getPayment returns a single item from the database.
func getPayment(hash string) (*Payment, error) {
	res := Payment{}
	if err := db.GetCollection("payment").Find(bson.M{"nodeHash": hash}).One(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

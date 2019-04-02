package item

import (
	"fmt"
	"log"

	"github.com/geo-graph-service/api/models/item/db"

	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Hash string `json:"hash" bson:"hash"`
}

func CreateNode(node *Node) error {
	_, err := findNode(node.Hash)
	if err != nil {
		return err
	} else {
		if err := db.SaveItem(node, "nodes"); err != nil {
			return err
		}
		return nil
	}
}

func DeleteNode(hash string) error {
	_, err := findNode(hash)
	if err != nil {
		log.Println("No node in db ", err)
		return err
	} else {
		if err := removeNode(hash, "nodes"); err != nil {
			return err
		}
		return nil
	}
}

//test
func findNode(hashNode string) (*Node, error) {
	var res = new(Node)
	err := db.GetCollection("nodes").Find(bson.M{"hash": hashNode}).One(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RemoveAll() error {
	res, err := getAllNodes()
	if err != nil {
		return fmt.Errorf("Cant load trustlines from db: %v", err)
	}
	for i := range res {
		err := removeNode(res[i].Hash, "node")
		if err != nil {
			log.Printf("range removing error occured: %v", err)
		}
	}
	return nil
}

func removeNode(hash string, tableName string) error {
	return db.GetCollection(tableName).Remove(bson.M{"hash": hash})
}

func getAllNodes() ([]Node, error) {
	res := []Node{}
	if err := db.GetCollection("nodes").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

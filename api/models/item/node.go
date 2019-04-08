package item

import (
	"fmt"
	"log"

	"github.com/GeoServer/project/api/models/item/db"

	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Hash string `json:"hash" bson:"hash"`
}

func CreateNode(node *Node) error {
	obj, err := FindNode(node.Hash)
	if obj != nil {
		return fmt.Errorf("Node is already exists in db: %v", err)
	}
	if err := db.SaveItem(node, "nodes"); err != nil {
		return fmt.Errorf("Problem with saving item in db: %v", err)
	}
	return nil
}

func DeleteNode(hash string) error {
	_, err := FindNode(hash)
	if err != nil {
		log.Println("No node in db: %v", err)
		return err
	}
	if err := removeNode(hash, "nodes"); err != nil {
		return err
	}
	return nil
}

func FindNode(hashNode string) (*Node, error) {
	var res = new(Node)
	err := db.GetCollection("nodes").Find(bson.M{"hash": hashNode}).One(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RemoveAllNodes() error {
	res, err := GetAllNodes()
	if err != nil {
		return fmt.Errorf("Can't load nodes from db: %v", err)
	}
	for i := range res {
		err := removeNode(res[i].Hash, "nodes")
		if err != nil {
			log.Printf("All nodes removing error: %v", err)
		}
	}
	return nil
}

func removeNode(hash string, tableName string) error {
	return db.GetCollection(tableName).Remove(bson.M{"hash": hash})
}

func GetAllNodes() ([]Node, error) {
	res := []Node{}
	if err := db.GetCollection("nodes").Find(nil).All(&res); err != nil {
		return nil, err
	}
	return res, nil
}

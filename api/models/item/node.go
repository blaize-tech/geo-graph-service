package item

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GeoServer/project/api/models/item/db"

	"gopkg.in/mgo.v2/bson"
)

type TopologyList struct {
	Nodes []Node `json:"nodes"`
}

type ListNodes struct {
	RangeFrom string `json:"rangefrom"`
	RangeTill string `json:"rangetill"`
	Split     string `json:"split"`
}

type Node struct {
	Hash        string         `json:"hash" bson:"hash"`
	State       string         `json:"-"bson:"state"`
	Date        time.Time      `json:"created" bson:"date"`
	OutGoingTLS []TrustlineAPI `json:"outgoing_tls,omitempty"`
}

func CreateNode(node *Node) error {
	node.State = "on"
	obj, err := FindNode(node.Hash, node.State)
	if obj != nil {
		return fmt.Errorf("Node is already exists in db: %v", err)
	}
	if err := db.SaveItem(node, "nodes"); err != nil {
		return fmt.Errorf("Problem with saving node in db: %v", err)
	}
	if err := db.SaveItem(node, "nodes_history"); err != nil {
		return fmt.Errorf("Problem with saving node_history in db: %v", err)
	}
	return nil
}

func DeleteNode(hash string) error {

	nod, err := FindNode(hash, "on")
	if err != nil {
		log.Println("No node in db: %v", err)
		return err
	}
	if err := removeNode(hash, "nodes"); err != nil {
		return err
	}

	nod.State = "off"
	nod.Date = time.Now()
	if err := db.SaveItem(nod, "nodes_history"); err != nil {
		return fmt.Errorf("Problem with saving node_history in db: %v", err)
	}
	return nil
}

func FindNode(hashNode, state string) (*Node, error) {
	var res = new(Node)
	err := db.GetCollection("nodes").Find(bson.M{"hash": hashNode, "state": state}).One(&res)
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
		res[i].State = "off"
		if err := db.SaveItem(res[i], "nodes_history"); err != nil {
			return fmt.Errorf("Problem with saving removing node_history in db: %v", err)
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

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func ActiveTopologyByDate(str string) (res TopologyList, err error) {
	strin := strings.Split(str, ".")
	year, _ := strconv.Atoi(strin[0])
	month, _ := strconv.Atoi(strin[1])
	day, _ := strconv.Atoi(strin[2])
	date := date(year, month, day)
	var nodesList []Node
	if err = db.GetCollection("nodes_history").Find(bson.M{"date": bson.M{"$lte": date}}).All(&nodesList); err != nil {
		return
	}
	var hash string
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for len(nodesList) != 0 {
			hash = nodesList[0].Hash
			node, err := getActiveNode(hash, nodesList)
			if err == nil {
				res.Nodes = append(res.Nodes, node)
			}
			nodesList = delNodes(node.Hash, nodesList)
		}
	}()
	var premadeTrustlines, trustlinesList []Trustline
	if err = db.GetCollection("trustline_history").Find(bson.M{"date": bson.M{"$lte": date}}).All(&trustlinesList); err != nil {
		return
	}
	go func() {
		for len(trustlinesList) != 0 {
			defer wg.Done()
			src, dst := trustlinesList[0].Source, trustlinesList[0].Destination
			trust, err := getActiveTrustline(src, dst, trustlinesList)
			if err == nil {
				premadeTrustlines = append(premadeTrustlines, trust)
			}
			trustlinesList = delTrustlines(src, dst, trustlinesList)
		}

	}()
	wg.Wait()
	for i, v := range res.Nodes {
		for _, val := range premadeTrustlines {
			if v.Hash == val.Source {
				res.Nodes[i].OutGoingTLS = append(v.OutGoingTLS, trustlineRepacker(val))

			}
		}
	}
	return
}

func delTrustlines(src, dst string, list []Trustline) []Trustline {
	for i := range list {
		if list[i].Source == src && list[i].Destination == dst {
			list = append(list[:i], list[i+1:]...)
			list = delTrustlines(src, dst, list)
			return list
		}
		continue
	}
	return list

}

func delNodes(hash string, list []Node) []Node {
	for i := range list {
		if list[i].Hash == hash {
			list = append(list[:i], list[i+1:]...)

			list = delNodes(hash, list)
			return list
		}
		continue
	}
	return list
}

func getActiveTrustline(src, dst string, trustlines []Trustline) (Trustline, error) {
	var activeTrust = Trustline{}
	for _, v := range trustlines {
		if v.Source == src && v.Destination == dst {
			activeTrust = v
		}
	}
	if activeTrust.State == "on" {

		return activeTrust, nil
	}

	return activeTrust, fmt.Errorf("Inactive")
}

func getActiveNode(hash string, nodes []Node) (Node, error) {
	var activeNode = Node{}
	for _, v := range nodes {
		if v.Hash == hash {
			activeNode = v
		}
	}
	if activeNode.State == "on" {

		return activeNode, nil
	}

	return activeNode, fmt.Errorf("Inactive")
}

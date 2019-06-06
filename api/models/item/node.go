package item

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"geo-graph-service/api/models/item/db"

	"gopkg.in/mgo.v2/bson"
)

type Range struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Count  int    `json:"count"`
}

type RangeResponse struct {
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Count  int       `json:"count"`
	Growth int       `json:"growth"`
}

type RangeResponseFull struct {
	Records int             `json:"records"`
	Stamps  []RangeResponse `json:"stamps"`
}

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
	State       string         `json:"-" bson:"state"`
	Date        time.Time      `json:"created" bson:"date"`
	OutGoingTLS []TrustlineAPI `json:"outgoing_tls,omitempty"`
}

var SwitchTypeStep = map[string]int{
	"day":   24,
	"week":  168,
	"month": 720,
}

const Trunc = 24 * time.Hour

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

func GetNodesByDate(date time.Time) (res TopologyList, err error) {
	var nodesList []Node
	if err = db.GetCollection("nodes_history").Find(bson.M{"date": bson.M{"$lte": date}}).All(&nodesList); err != nil {
		log.Printf("DB error: GetNodesByDate failed cause %v", err)
		return
	}
	var hash string
	for len(nodesList) != 0 {
		hash = nodesList[0].Hash
		node, err := getActiveNode(hash, nodesList)
		if err == nil {
			res.Nodes = append(res.Nodes, node)
		}
		nodesList = delNodes(node.Hash, nodesList)
	}
	return res, nil
}

func GetTrustlinesByDate(date time.Time) (premadeTrustlines []Trustline, err error) {
	var trustlinesList []Trustline
	if err = db.GetCollection("trustline_history").Find(bson.M{"date": bson.M{"$lte": date}}).All(&trustlinesList); err != nil {
		return
	}
	for len(trustlinesList) != 0 {
		src, dst := trustlinesList[0].Source, trustlinesList[0].Destination
		trust, err := getActiveTrustline(src, dst, trustlinesList)
		if err == nil {
			premadeTrustlines = append(premadeTrustlines, trust)
		}
		trustlinesList = delTrustlines(src, dst, trustlinesList)
	}
	return premadeTrustlines, nil
}

func TopologyRepack(str string) (res TopologyList, err error) {
	strin := strings.Split(str, ".")
	year, _ := strconv.Atoi(strin[0])
	month, _ := strconv.Atoi(strin[1])
	day, _ := strconv.Atoi(strin[2])
	date := date(year, month, day)

	res, _ = GetNodesByDate(date)
	premadeTrustlines, _ := GetTrustlinesByDate(date)
	for i, v := range res.Nodes {
		for _, val := range premadeTrustlines {
			if v.Hash == val.Source {
				res.Nodes[i].OutGoingTLS = append(v.OutGoingTLS, TrustlineRepacker(val))

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

func RangeList(rng Range) (resFull RangeResponseFull, err error) {
	step, ok := SwitchTypeStep[rng.Type]
	if !ok {
		return resFull, fmt.Errorf("Invalid type filter!\n\t Should be 'day/week/month'")
	}

	var nodeActual, nodeLast Node
	if err = db.GetCollection("nodes_history").Find(nil).One(&nodeLast); err != nil {
		return
	}

	dbSize, err := db.GetCollection("nodes_history").Count()
	if err != nil {
		return
	}
	if err = db.GetCollection("nodes_history").Find(nil).Skip(dbSize - 1).One(&nodeActual); err != nil {
		return
	}

	if rng.Offset == 0 && rng.Count == 0 {
		ls := nodeLast.Date.Truncate(time.Hour * 24)
		resFull = returnAllRange(nodeActual.Date, ls, time.Duration(step))
		return
	}

	if rng.Count == 0 {
		actual := nodeActual.Date.Add(-time.Hour * time.Duration(step*rng.Offset))
		ls := nodeLast.Date.Truncate(time.Hour * 24)
		if ls.After(actual) {
			return resFull, fmt.Errorf("Date is out of actual")
		}
		resFull = returnAllRange(actual, ls, time.Duration(step))
		return
	}

	actual := nodeActual.Date.Add(-time.Hour * time.Duration(rng.Offset*step))
	last := actual.Add(-time.Hour * time.Duration(step*rng.Count))

	if nodeLast.Date.After(actual) {
		return resFull, fmt.Errorf("Date is out of actual")
	}

	if nodeLast.Date.After(last) && !nodeLast.Date.After(actual) {
		last = nodeLast.Date
	}

	var k int
	for i := 0; i < rng.Count; i++ {
		if last.After(actual) && !last.Add(-time.Hour*time.Duration(step)).After(actual) {
			if i == 0 {
				var mck RangeResponse
				resFull.Stamps = append(resFull.Stamps, mck)
				k = 0
			} else {
				k = i - 1
			}
			a, _ := GetNodesByDate(actual)
			prevCount, _ := GetNodesByDate(last.Add(-time.Hour * time.Duration(step)))
			resFull.Stamps[k].End = actual
			resFull.Stamps[k].Count = len(a.Nodes)
			resFull.Stamps[k].Growth = len(a.Nodes) - len(prevCount.Nodes)
			break
		}

		b, _ := GetNodesByDate(last)
		add := rangePack(last, len(b.Nodes), time.Duration(step))
		resFull.Stamps = append(resFull.Stamps, add)
		last = last.Add(time.Hour * time.Duration(step))
	}
	resFull.Records = len(resFull.Stamps)
	resFull.Stamps = revers(resFull.Stamps)
	return
}

func rangePack(start time.Time, countPrev int, step time.Duration) (res RangeResponse) {
	res.Start = start
	res.End = start.Add(time.Hour * step)
	c, _ := GetNodesByDate(res.End)
	res.Count = len(c.Nodes)
	res.Growth = res.Count - countPrev
	return
}

func returnAllRange(actual, last time.Time, step time.Duration) (res RangeResponseFull) {
	for i := 0; ; i++ {
		var k int
		if last.After(actual) && !last.Add(-time.Hour*step).After(actual) {
			a, _ := GetNodesByDate(actual)
			if i == 0 {
				var mck RangeResponse
				res.Stamps = append(res.Stamps, mck)
				k = 0
			} else {
				k = i - 1
			}
			prevCount, _ := GetNodesByDate(last.Add(-time.Hour * step))
			res.Stamps[k].End = actual
			res.Stamps[k].Count = len(a.Nodes)
			res.Stamps[k].Growth = len(a.Nodes) - len(prevCount.Nodes)
			break
		}
		b, _ := GetNodesByDate(last)
		add := rangePack(last, len(b.Nodes), step)
		res.Stamps = append(res.Stamps, add)
		last = last.Add(time.Hour * time.Duration(step))
	}

	res.Records = len(res.Stamps)
	res.Stamps = revers(res.Stamps)
	return
}

func revers(a []RangeResponse) []RangeResponse {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/GeoServer/project/api/models/item"
)

//ClientMock for client vizualization
type ClientMock struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Delete      bool
	Time        time.Time
}

func handleError(err error, message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(message, err)))
}

//
func Topology(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("date") {
	case "":
		http.Error(w, "Please send a correct url key body", 400)
		return
	default:
		res, err := item.TopologyRepack(r.FormValue("date"))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Topology return failed:%v", err)))
			return
		}
		out, _ := json.Marshal(res)
		w.Write(out)
		return
	}
}

//
func TopologyRange(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.FormValue("type") == "" || r.FormValue("offset") == "" || r.FormValue("count") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}
	var rng item.Range
	rng.Type = r.FormValue("type")
	rng.Offset, err = strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		handleError(err, "Unable to parse offset: %v", w)
		return
	}
	rng.Count, err = strconv.Atoi(r.FormValue("count"))
	if err != nil {
		handleError(err, "Unable to parse count: %v", w)
		return
	}
	rng.Type = r.FormValue("type")

	res, err := item.RangeList(rng)
	if err != nil {
		handleError(err, "Topology range err: %v", w)
		return
	}
	resJSON, _ := json.Marshal(res)
	w.Write(resJSON)
	return
}

//CreateNode creates request node to the database
func CreateNode(s *Server, w http.ResponseWriter, req *http.Request) {
	var nod = new(item.Node)
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&nod)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	nod.Date = time.Now()
	nod.State = "on"
	if err = item.CreateNode(nod); err != nil {
		handleError(err, "Failed to insert node data in db: %v", w)
		return
	}
	nd := ClientMock{Source: nod.Hash, Destination: nod.Hash, Time: time.Now()}
	bs, _ := json.Marshal(nd)
	s.pushEvent(bs)
	w.Write([]byte("OK"))
	return
}

//DeleteNode deletes request node from database
func DeleteNode(s *Server, w http.ResponseWriter, req *http.Request) {
	switch req.FormValue("hash") {
	case "":
		http.Error(w, "Please send a correct url key body", 400)
		return
	default:
		if err := item.DeleteNode(req.FormValue("hash")); err != nil {
			handleError(err, "Failed to delete node from db: %v", w)
			return
		}
		nd := ClientMock{Source: req.FormValue("hash"), Destination: req.FormValue("hash"), Time: time.Now(), Delete: true}
		bs, _ := json.Marshal(nd)
		s.pushEvent(bs)
		w.Write([]byte("OK"))
		return
	}
}

//PostTrustline creates trustline by request and adds it to the database
func PostTrustline(s *Server, w http.ResponseWriter, req *http.Request) {
	trustline := item.Trustline{}
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&trustline)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = item.FindNode(trustline.Destination, "on")
	if err != nil {
		node := item.Node{Hash: trustline.Destination}
		item.CreateNode(&node)
		nd := ClientMock{Source: node.Hash, Destination: node.Hash, Time: time.Now()}
		bs, _ := json.Marshal(nd)
		s.pushEvent(bs)
	}

	_, err = item.FindNode(trustline.Source, "on")
	if err != nil {
		node := item.Node{Hash: trustline.Source}
		item.CreateNode(&node)
		nd := ClientMock{Source: node.Hash, Destination: node.Hash, Time: time.Now()}
		bs, _ := json.Marshal(nd)
		s.pushEvent(bs)
	}

	//set current time
	trustline.Time = time.Now()
	err = item.PostTrustline(&trustline)
	if err != nil {
		handleError(err, "Cannot create trustline, err: %v ", w)
		return
	}
	//write bytes to event
	bs, _ := json.Marshal(trustline)
	s.pushEvent(bs)
	w.Write([]byte("OK"))
	return
}

//DeleteTrustline removes request trustline from database
func DeleteTrustline(s *Server, w http.ResponseWriter, req *http.Request) {
	if req.FormValue("src") == "" || req.FormValue("dst") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}

	if _, err := item.FindTrustline(req.FormValue("src"), req.FormValue("dst")); err != nil {
		err := item.DeleteTrustline(req.FormValue("src"), req.FormValue("dst"))
		if err != nil {
			handleError(err, "Cannont delete trustline: %v", w)
		}
		nd := ClientMock{Source: req.FormValue("src"), Destination: req.FormValue("dst"), Time: time.Now(), Delete: true}

		bs, _ := json.Marshal(nd)
		s.pushEvent(bs)
		w.Write([]byte("OK"))
		return
	}
	w.Write([]byte("No trustline to delete"))
	return
}

//PostPaymentItem adds payment to service
func PostPaymentItem(s *Server, w http.ResponseWriter, req *http.Request) {
	payment := item.Payment{}
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	err := json.NewDecoder(req.Body).Decode(&payment)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	payment.Time = time.Now()
	for i := range payment.Paths {
		src := []string{payment.Source}
		payment.Paths[i] = append(src, payment.Paths[i]...)
		payment.Paths[i] = append(payment.Paths[i], payment.Destination)
	}
	bs, err := json.Marshal(payment)
	s.pushEvent(bs)

	w.Write([]byte("OK"))
}

//GetItems return all items data from database
func GetItems() ([]item.Node, []item.Trustline, []item.Payment) {

	rsNodes, err := item.GetAllNodes()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsNodes = nil
	}

	rsTrustlines, err := item.GetAllTrustlines()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsTrustlines = nil
	}

	rsPayments, err := item.GetAllPayments()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsPayments = nil
	}

	return rsNodes, rsTrustlines, rsPayments
}

// DeleteAll removes a single item (identified by parameter) from the database.
func DeleteAll(s *Server, w http.ResponseWriter, req *http.Request) error {
	switch req.FormValue("key") {
	case "":
		err := errors.New("no URL key")
		return err
	case getConfig().Key:
		var wg sync.WaitGroup
		wg.Add(2)
		nodes, trustlines, _ := GetItems()
		if err := item.RemoveAllNodes(); err != nil {
			handleError(err, "Deleting nodes failed: %v", w)
			return err
		}
		if err := item.RemoveAllTrustlines(); err != nil {
			handleError(err, "Deleting trustlines failed: %v", w)
			return err
		}

		go func() {
			for _, i := range trustlines {
				nd := ClientMock{Source: i.Source, Destination: i.Destination, Time: time.Now(), Delete: true}
				bs, _ := json.Marshal(nd)
				s.pushEvent(bs)
			}
			wg.Done()
		}()

		go func() {
			for _, i := range nodes {
				nd := ClientMock{Source: i.Hash, Destination: i.Hash, Time: time.Now(), Delete: true}
				bs, _ := json.Marshal(nd)
				s.pushEvent(bs)
			}
			wg.Done()
		}()

		wg.Wait()
		w.Write([]byte("OK"))

	default:
		http.Error(w, "Invalid key.", 405)
		return fmt.Errorf("Error: key value")
	}
	return nil
}

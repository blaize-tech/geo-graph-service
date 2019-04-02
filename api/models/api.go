package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/geo-graph-service/api/models/item"
)

func handleError(err error, message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(message, err)))
}

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
	if err = item.CreateNode(nod); err != nil {
		handleError(err, "Failed to insert node data in db: %v", w)
		return
	} else {
		w.Write([]byte("OK"))
		return
	}
}

func DeleteNode(s *Server, w http.ResponseWriter, req *http.Request) {
	switch req.FormValue("hash") {
	case "":
		http.Error(w, "Please send a correct url key body", 400)
		return
	default:
		if err := Item.DeleteNode(req.FormValue("hash")); err != nil {
			handleError(err, "Failed to delete node from db: %v", w)
			return
		} else {
			w.Write([]byte("OK"))
		}

	}
	return
}

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
	//set current time
	trustline.Time = time.Now()
	err = item.PostTrustline(trustline)
	if err != nil {
		handleError(err, "Cannot create trustline ", w)
	}
	//write bytes to event
	bs, _ := json.Marshal(trustline)
	s.pushEvent(bs)
	w.Write([]byte("OK"))
	return
}

func DeleteTrustlineItem(s *Server, w http.ResponseWriter, req *http.Request) {

	if req.FormValue("src") == "" || req.FormValue("dst") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}

	err := item.DeleteTrustline(req.FormValue("src"), req.FormValue("dst"))
	if err != nil {
		handleError(err, "Cannont delete trustline,w")
	} //write bytes to event
	// bs, _ := json.Marshal(trustline)
	// s.pushEvent(bs)

	// w.Write([]byte("OK"))
	return
}

func PostPaymentItem(s *Server, payment *Payment) {

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
	bs, err := json.Marshal(payment)
	s.pushEvent(bs)

	w.Write([]byte("OK"))
}

func GetItems() ([]Node, []Trustline, []Payment) {
	rsTrustlines, err := getAllTrustlines()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsTrustlines = nil
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println("Failed to load database items:", err)
	}

	return nil, rsTrustlines, rsPayments
}

// DeleteItem removes a single item (identified by parameter) from the database.
func DeleteAll(w http.ResponseWriter, req *http.Request) error {
	switch req.FormValue("key") {
	case "":
		err := errors.New("No URL key")
		return err
	case getConfig().Key:
		if err := item.RemoveAllTrustlines(); err != nil {
			handleError(err, "Deleteting trustlines failed", w)
		}
		if err := item.RemoveAllNodes(); err != nil {
			handleError(err, "Deleteting nodes failed", w)
		}
	default:
		http.Error(w, "Invalid key.", 405)
	}
	w.Write([]byte("OK"))
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func handleError(err error, message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(message, err)))
}

func geItems() ([]Trustline, []Payment){
	rsTrustlines, err := getAllTrustlines()
	if err != nil {
		log.Println( "Failed to load database items:", err)
		rsTrustlines = nil
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println( "Failed to load database items:", err)
	}

	return rsTrustlines, rsPayments
}

// PostItem saves an item (form data) into the database.
func postTrustlineItem(s *Server, w http.ResponseWriter, req *http.Request) {
	var trustline = new(Trustline)
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

	//if another node empty create it
	if trustline.Destination != trustline.Source {
		_, err = getTrustline(trustline.Destination)
		if err != nil {
			newTrustline := Trustline{trustline.Destination, trustline.Destination, false, trustline.Time}
			if err := saveItem(newTrustline, "trustline"); err != nil {
				handleError(err, "Failed to save data of Destination: %v", w)
				return
			} else {
				//write bytes to event
				bs, _ := json.Marshal(newTrustline)
				s.pushEvent(bs)
			}
		}
	}

	rt, err := getTrustline(trustline.Source)
	if err != nil {
		if err := saveItem(trustline, "trustline"); err != nil {
			handleError(err, "Failed to save data of Source: %v", w)
			return
		}
	} else {
		if rt.Destination != trustline.Destination {
			if err := saveItem(trustline, "trustline"); err != nil {
				handleError(err, "Failed to save data: %v", w)
				return
			}
		} else {
			if err := updateTrustline(trustline); err != nil {
				handleError(err, "Failed to update data: %v", w)
				return
			}
		}
	}

	//write bytes to event
	bs, _ := json.Marshal(trustline)
	s.pushEvent(bs)

	w.Write([]byte("OK"))
}

func deleteTrustlineItem(s *Server, w http.ResponseWriter, req *http.Request) {
	var trustline = new(Trustline)
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&trustline)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	errFirst := removeItem(trustline.Source, trustline.Destination, "trustline");
	errSecond := removeItem(trustline.Destination, trustline.Source, "trustline");

	if  errFirst != nil && errSecond != nil{
		handleError(err, "Failed to remove data: %v", w)
		return
	}

	trustline.Delete = true

	//write bytes to event
	bs, err := json.Marshal(trustline)
	s.pushEvent(bs)

	w.Write([]byte("OK"))
}


func postPaymentItem(s *Server, w http.ResponseWriter, req *http.Request) {

	var payment = new(Payment)
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	err := json.NewDecoder(req.Body).Decode(&payment)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	bs, err := json.Marshal(payment)
	s.pushEvent(bs)

	payment.Time = time.Now()

	w.Write([]byte("OK"))
}

// DeleteItem removes a single item (identified by parameter) from the database.
func deleteItem(w http.ResponseWriter, req *http.Request) {
	if err := clearAll(); err != nil {
		log.Println( "Failed to save data: %v", err)
		return
	}
	w.Write([]byte("OK"))
}
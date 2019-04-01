package main

import (
	"encoding/json"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"
)

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

	//no route node, no trustline
	_, err = findNode(trustline.Destination)
	if err != nil {
		handleError(err, "Failed to create trustline, destination not found: %v", w)
		return
	}
	_, err = findNode(trustline.Source)
	if err != nil {
		handleError(err, "Failed to create trustline, source not found: %v", w)
		return
	}

	_, err = findTrustline(trustline.Destination, trustline.Source)
	if err != nil {
		trustline.Equivalent = equi
		if err := saveItem(trustline, "trustline"); err != nil {
			handleError(err, "Failed to create node: %v", w)
			return
		}
		//write bytes to event
		bs, _ := json.Marshal(trustline)

		s.pushEvent(bs)

		w.Write([]byte("OK"))
	} else {
		handleError(err, "Cant create node: %v", w)
		return
	}
}

func deleteTrustlineItem(s *Server, w http.ResponseWriter, req *http.Request) {

	// var trustline = new(Trustline)

	// err := json.NewDecoder(req.Body).Decode(&trustline)
	// if err != nil {
	// 	http.Error(w, err.Error(), 400)
	// 	return
	// }
	//log.Print(req.FormValue("src"), req.FormValue("dst"))
	if req.FormValue("src") == "" || req.FormValue("dst") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}

	_, err := findTrustline(req.FormValue("src"), req.FormValue("dst"))
	//log.Println(err)
	if err == nil {
		if err = removeItem(req.FormValue("src"), req.FormValue("dst"), "trustline"); err != nil {
			if err = removeItem(req.FormValue("dst"), req.FormValue("src"), "trustline"); err != nil {
				handleError(err, "Failed to remove trustline: %v", w)
				return
			}

		}
		trustline := Trustline{Source: req.FormValue("src"), Destination: req.FormValue("dst"), Equivalent: equi, Time: time.Now()}
		//write bytes to event
		bs, _ := json.Marshal(trustline)
		s.pushEvent(bs)

		w.Write([]byte("OK"))
		return
	} else {
		handleError(err, "Error cant remove node: %v", w)
	}

}

func findTrustline(hashSource string, hashDestination string) (*Trustline, error) {
	res := Trustline{}
	if err := getCollection("trustline").Find(bson.M{"nodeSource": hashSource, "nodeDestination": hashDestination}).One(&res); err != nil {
		if err = getCollection("trustline").Find(bson.M{"nodeSource": hashDestination, "nodeDestination": hashSource}).One(&res); err != nil {
			return nil, err
		}
		return nil, err
	}
	return &res, nil
}

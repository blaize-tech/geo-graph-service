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

// create node
func postNodeItem(s *Server, w http.ResponseWriter, req *http.Request) {
	var node = new(Trustline)
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(req.Body).Decode(&node)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if node.Source == node.Destination {
		node.Time = time.Now()
		_, err = findTrustline(node.Source, node.Destination)
		if err != nil {
			if err := saveItem(node, "trustline"); err != nil {
				handleError(err, "Failed to save data of Source: %v", w)
				return
			}
		} else {
			if err := updateTrustline(node); err != nil {
				handleError(err, "Failed to update data: %v", w)
				return
			}
		}

		//write bytes to event
		bs, _ := json.Marshal(node)
		s.pushEvent(bs)

		w.Write([]byte("OK"))
	} else {
		handleError(err, "Cant create trustline: %v", w)
		return
	}
}

// delete node
func deleteNodesItem(s *Server, w http.ResponseWriter, req *http.Request) {
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
	if trustline.Source == trustline.Destination {
		rts, err := getTrustlinesBySource(trustline.Source)
		rtd, err := getTrustlinesByDestination(trustline.Destination)

		if err != nil || len(rtd)<=1 || len(rts)<=1 {
			err := removeItem(trustline.Source, trustline.Destination, "trustline");
			if  err != nil {
				handleError(err, "Failed to remove noda: %v", w)
				return
			} else {
				trustline.Delete = true

				//write bytes to event
				bs, _ := json.Marshal(trustline)
				s.pushEvent(bs)
				w.Write([]byte("OK"))
			}
		} else {
			handleError(err, "Failed to remove noda the trusline conected: %v", w)
			return
		}

	} else {
		handleError(err, "Error cant remove trustline: %v", w)
	}
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

	//if another nodes empty create it
	if trustline.Destination != trustline.Source {
		_, err := getTrustline(trustline.Source)
		if err != nil {
			newTrustline := Trustline{trustline.Source, trustline.Source, false, trustline.Time}
			if err := saveItem(newTrustline, "trustline"); err != nil {
				handleError(err, "Failed to create node: %v", w)
				return
			}
		}
		_, err = getTrustline(trustline.Destination)
		if err != nil {
			newTrustline := Trustline{trustline.Destination, trustline.Destination, false, trustline.Time}
			if err := saveItem(newTrustline, "trustline"); err != nil {
				handleError(err, "Failed to create node of Destination: %v", w)
				return
			} else {
				//write bytes to event
				bs, _ := json.Marshal(newTrustline)
				s.pushEvent(bs)
			}
		}

		_, errD := findTrustline(trustline.Destination, trustline.Source)
		_, errS := findTrustline(trustline.Source, trustline.Destination)

		if errS != nil && errD != nil {
			if err := saveItem(trustline, "trustline"); err != nil {
				handleError(err, "Failed to save data of Source: %v", w)
				return
			}
		} else {
			if err := updateTrustline(trustline); err != nil {
				handleError(err, "Failed to update data: %v", w)
				return
			}
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

	if trustline.Source != trustline.Destination {
		var errFirst, errSecond error
		_, errS := findTrustline(trustline.Source, trustline.Destination)
		if errS == nil {
			errFirst = removeItem(trustline.Source, trustline.Destination, "trustline");
		}
		_, errD := findTrustline(trustline.Source, trustline.Destination)
		if errD == nil {
			errSecond = removeItem(trustline.Destination, trustline.Source, "trustline");
		}
		if errFirst != nil && errSecond != nil {
			handleError(err, "Failed to remove data: %v", w)
			return
		}

		trustline.Delete = true

		//write bytes to event
		bs, _ := json.Marshal(trustline)
		s.pushEvent(bs)

		w.Write([]byte("OK"))
	} else {
		handleError(err, "Error cant remove node: %v", w)
	}

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
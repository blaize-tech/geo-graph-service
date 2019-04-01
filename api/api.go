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

func geItems() ([]Trustline, []Payment) {
	rsTrustlines, err := getAllTrustlines()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsTrustlines = nil
	}

	rsPayments, err := getAllPayments()
	if err != nil {
		log.Println("Failed to load database items:", err)
	}

	return rsTrustlines, rsPayments
}

func createNode1(s *Server, w http.ResponseWriter, req *http.Request) {
	var node = new(Node)
	if req.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	err := json.NewDecoder(req.Body).Decode(&node)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_, err = findNode(node.Hash)
	if err != nil {
		if err := saveItem(node, "nodes"); err != nil {
			handleError(err, "Failed to insert node data in db: %v", w)
			return
		}
		w.Write([]byte("OK"))
		return
	} else {
		w.Write([]byte("Node is already exists!"))
		return
	}
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

		if err != nil || len(rtd) <= 1 || len(rts) <= 1 {
			err := removeItem(trustline.Source, trustline.Destination, "trustline")
			if err != nil {
				handleError(err, "Failed to remove noda: %v", w)
				return
			} else {
				//	trustline.Delete = true

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

/*
// delete node by DELETE in url
func deleteNodesItem(s *Server, w http.ResponseWriter, req *http.Request) {
	var trustline = new(Trustline)
	node := req.FormValue("node")
	if node == "" {
		http.Error(w, "Please send a correct url parametrs", 400)
		return
	}
	trustline.Source = node
	trustline.Destination = node

		rts, err := getTrustlinesBySource(trustline.Source)
		rtd, err := getTrustlinesByDestination(trustline.Destination)

		if err != nil || len(rtd) <= 1 || len(rts) <= 1 {
			err := removeItem(trustline.Source, trustline.Destination, "trustline")
			if err != nil {
				handleError(err, "Failed to remove noda: %v", w)
				return
			} else {
				//	trustline.Delete = true

				//write bytes to event
				bs, _ := json.Marshal(trustline)
				s.pushEvent(bs)
				w.Write([]byte("OK"))
			}
		} else {
			handleError(err, "Failed to remove noda the trusline conected: %v", w)
			return
		}


}
*/

///////////////////////////////////////////////////
/*
func addTrustline(s *Server, w http.ResponseWriter, req *http.Request) {
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

	if _, err := findNode(trustline.Destination); err != nil {
		handleError(err, "Node doesnt exists", w)
		return
	}

	if _, err := findNode(trustline.Source); err != nil {
		handleError(err, "Node doesnt exists", w)
		return
	}

	if _, err := findTrustline(trustline.Destination,trustline.Source);


	if trustline.Destination != trustline.Source {



		//write bytes to event
		bs, _ := json.Marshal(trustline)
		s.pushEvent(bs)

		w.Write([]byte("OK"))
	} else {
		handleError(err, "Cant create node: %v", w)
		return
	}

}
*/
/////////////////////////////////////////////////////////

// DeleteItem removes a single item (identified by parameter) from the database.
func deleteItem(w http.ResponseWriter, req *http.Request) {
	if err := clearAll(); err != nil {
		log.Printf("Failed to save data: %v", err)
		return
	}
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

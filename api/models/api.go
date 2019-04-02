package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GeoServer/project/api/models/item"
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
	}
	w.Write([]byte("OK"))
	return

}

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
		w.Write([]byte("OK"))
		return
	}
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
	err = item.PostTrustline(&trustline)
	if err != nil {
		handleError(err, "Cannot create trustline,err: %v ", w)
		return
	}
	//write bytes to event
	bs, _ := json.Marshal(trustline)
	s.pushEvent(bs)
	w.Write([]byte("OK"))
	return
}

func DeleteTrustline(s *Server, w http.ResponseWriter, req *http.Request) {

	if req.FormValue("src") == "" || req.FormValue("dst") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}

	if i, _ := item.FindTrustline(req.FormValue("src"), req.FormValue("dst")); i == nil {
		err := item.DeleteTrustline(req.FormValue("src"), req.FormValue("dst"))
		if err != nil {
			handleError(err, "Cannont delete trustline", w)
		}
		w.Write([]byte("OK"))
		return
	}

	//write bytes to event
	// bs, _ := json.Marshal(trustline)
	// s.pushEvent(bs)

	w.Write([]byte("No trustline to delete"))
	return
}

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
	bs, err := json.Marshal(payment)
	s.pushEvent(bs)

	w.Write([]byte("OK"))
}

func GetItems() ([]item.Node, []item.Trustline, []item.Payment) {
	rsTrustlines, err := item.GetAllTrustlines()
	if err != nil {
		log.Println("Failed to load database items:", err)
		rsTrustlines = nil
	}

	rsPayments, err := item.GetAllPayments()
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
			return err
		}
		if err := item.RemoveAllNodes(); err != nil {
			handleError(err, "Deleting nodes failed", w)
			return err
		}
		w.Write([]byte("OK"))
	default:
		http.Error(w, "Invalid key.", 405)
		return fmt.Errorf("Error :key value")
	}
	return nil
}

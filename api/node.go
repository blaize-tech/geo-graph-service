package main

import (
	"encoding/json"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

func createNode(s *Server, w http.ResponseWriter, req *http.Request) {
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

func deleteNode(s *Server, w http.ResponseWriter, req *http.Request) {
	if req.FormValue("hash") == "" {
		http.Error(w, "Please send a correct url key body", 400)
		return
	}

	_, err := findNode(req.FormValue("hash"))
	if err == nil {
		if err := removeNode(req.FormValue("hash"), "nodes"); err != nil {
			handleError(err, "Failed to remove node data from db: %v", w)
			return
		}
		w.Write([]byte("OK"))
		return
	} else {
		w.Write([]byte("No Node in the db !"))
		return
	}
}

//test
func findNode(hashNode string) (*Node, error) {
	var res = new(Node)
	err := getCollection("nodes").Find(bson.M{"hash": hashNode}).One(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//test

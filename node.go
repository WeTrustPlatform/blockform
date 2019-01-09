package main

import (
	"net/http"

	"github.com/WeTrustPlatform/blockform/model"
	"goji.io/pat"
)

func handleNode(w http.ResponseWriter, r *http.Request) {
	id := pat.Param(r, "id")
	http.Redirect(w, r, "/node/"+id+"/general", http.StatusSeeOther)
}

func handleNodeTab(w http.ResponseWriter, r *http.Request) {
	id := pat.Param(r, "id")
	tab := pat.Param(r, "tab")
	node := model.Node{}
	db.Find(&node, id)
	tmpl.ExecuteTemplate(w, "node_"+tab+".html", struct {
		Tab  string
		Node model.Node
	}{
		tab,
		node,
	})
}

func handleNodeExplorer(w http.ResponseWriter, r *http.Request) {
	nodeID := pat.Param(r, "nodeid")
	tab := "explorer"
	class := pat.Param(r, "class")
	id := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, nodeID)
	tmpl.ExecuteTemplate(w, "node_"+tab+"_"+class+".html", struct {
		Tab  string
		ID   string
		Node model.Node
	}{
		tab,
		id,
		node,
	})
}

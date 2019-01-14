package main

import (
	"net/http"

	"github.com/WeTrustPlatform/blockform/model"
	"goji.io/pat"
)

// To try this:
// curl -X POST -F 'title=Geth has been restarted' -F 'type=issue' -F 'description=The blocknumber was lagging behind etherscan for more than 200 blocks' http://localhost:3000/node/:id/event/:apikey

func handleEvent(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	APIKey := pat.Param(r, "apikey")

	node := model.Node{}
	db.Find(&node, ID)

	if node.APIKey != APIKey {
		w.WriteHeader(403)
		return
	}

	title := r.FormValue("title")
	type0 := r.FormValue("type")
	description := r.FormValue("description")
	if title == "" || (type0 != model.Issue && type0 != model.Fine) || description == "" {
		w.WriteHeader(400)
		return
	}

	db.Create(&model.Event{
		NodeID:      node.ID,
		Type:        type0,
		Title:       title,
		Description: description,
	})
}

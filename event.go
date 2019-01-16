package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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

	event := model.Event{
		NodeID:      node.ID,
		Type:        type0,
		Title:       title,
		Description: description,
	}
	db.Create(&event)

	go notifySlack(node, event)
}

func notifySlack(node model.Node, event model.Event) {
	url := os.Getenv("SLACK_HOOK")

	// only notify issues, and only of the SLACK_HOOK is set
	if event.Type != model.Issue || url == "" {
		return
	}

	nodeURL := fmt.Sprintf("%s/node/%d/activity", os.Getenv("SITE_URL"), node.ID)

	json := []byte(`{
	"attachments": [
		{
			"pretext": "On node ` + node.Name + `",
			"title": "` + event.Title + `",
			"title_link": "` + nodeURL + `",
			"text": "` + event.Description + `",
			"color": "#dc3545"
		}
	]
}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))
}

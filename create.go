package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/WeTrustPlatform/blockform/model"
	"github.com/sethvargo/go-password/password"
)

func handleCreateForm(w http.ResponseWriter, r *http.Request) {
	var keys []string
	for k := range providers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	tmpl.ExecuteTemplate(w, "create.html", keys)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(500)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = "bf-" + password.MustGenerate(8, 4, 0, true, false)
	}
	provider := r.FormValue("provider")
	if provider == "" {
		provider = "aws"
	}
	networkID, _ := strconv.ParseInt(r.FormValue("network_id"), 10, 64)
	if networkID == 0 {
		rand.Seed(time.Now().UnixNano())
		networkID = int64(rand.Intn(65536))
	}
	syncMode := r.FormValue("sync_mode")
	if syncMode == "" {
		syncMode = model.Fast
	}
	APIKey := password.MustGenerate(8, 4, 0, false, false)

	node := model.Node{
		Name:          name,
		CloudProvider: provider,
		DomainName:    r.FormValue("domain_name"),
		NetworkType:   r.FormValue("network_type"),
		NetworkID:     uint64(networkID),
		SyncMode:      syncMode,
		APIKey:        APIKey,
		Status:        model.Creating,
	}

	db.Create(&node)

	cloud := providers[node.CloudProvider]
	wg.Add(1)
	go func() {
		defer wg.Done()
		cloud.CreateNode(context.Background(), node,
			// On Success
			func(VMID, DomainName string) {
				node.Status = model.Deployed
				node.VMID = VMID
				node.DomainName = DomainName
				db.Save(&node)
				db.Create(&model.Event{
					NodeID: node.ID,
					Title:  "Node successfully created",
					Type:   model.Fine,
				})
				log.Println("Done creating node " + node.Name)
			},
			// On Error
			func(err error) {
				node.Status = model.Error
				db.Save(&node)
				db.Create(model.Event{
					NodeID:      node.ID,
					Title:       "Error when creating node",
					Description: err.Error(),
					Type:        model.Fine,
				})
				log.Println("Error while create node ", node.Name, err)
			},
		)
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/WeTrustPlatform/blockform/model"
	"goji.io/pat"
)

func handleNodeDelete(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, ID)

	db.Model(&model.Node{}).Where("id=?", ID).Update("Status", model.Deleting)

	cloud := providers[node.CloudProvider]
	log.Println("Deleting node", node.Name)
	if node.Status == model.Error {
		db.Where("id=?", ID).Delete(&model.Node{})
		log.Println("Done deleting node " + node.Name)
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cloud.DeleteNode(context.Background(), node,
				// On Success
				func() {
					db.Where("id=?", ID).Delete(&model.Node{})
					log.Println("Done deleting node " + node.Name)
				},
				// On Error
				func(err error) {
					db.Model(&model.Node{}).Where("id=?", ID).Update("Status", model.Deployed)
					log.Println("Error while deleting node", node.Name, err)
				},
			)
		}()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

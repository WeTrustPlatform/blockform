package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/WeTrustPlatform/blockform/model"
	"github.com/WeTrustPlatform/blockform/sshcmd"
	"goji.io/pat"
)

// RebootNode reboots the VM where the node is hosted
func rebootNode(ctx context.Context, node model.Node, callback func()) {
	_, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		"sudo reboot",
	)
	if err != nil {
		log.Println(err)
	}
	callback()
}

func handleReboot(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, ID)
	log.Println("Rebooting node", node.Name)
	go rebootNode(context.Background(), node, func() {
		db.Create(&model.Event{
			NodeID: node.ID,
			Type:   model.Fine,
			Title:  "The node has been manually rebooted",
		})
		log.Println("Done rebooting node " + node.Name)
	})
	http.Redirect(w, r, "/node/"+ID+"/activity", http.StatusSeeOther)
}

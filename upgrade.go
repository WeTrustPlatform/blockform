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

// UpgradeNode upgrades the geth to latest version
func upgradeNode(ctx context.Context, node model.Node, callback func()) {
	_, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		`git clone https://github.com/WeTrustPlatform/manage-node-scripts.git \
		&& cd manage-node-scripts \
		&& ./download-geth.sh \
		&& sudo systemctl stop geth \
		&& sudo cp geth /usr/bin/ \
		&& cd .. && rm -rf manage-node-scripts`,
	)
	if err != nil {
		log.Println(err)
	}

	// no matter if the process success or failure
	// restart geth
	_, _, err = sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		"sudo systemctl restart geth",
	)
	if err != nil {
		log.Println(err)
	}

	callback()
}

func handleUpgrade(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, ID)
	log.Println("Upgrading geth", node.Name)
	go upgradeNode(context.Background(), node, func() {
		db.Create(&model.Event{
			NodeID: node.ID,
			Type:   model.Fine,
			Title:  "The node has been upgraded to the latest geth",
		})
		log.Println("Done upgrading node " + node.Name)
	})
	http.Redirect(w, r, "/node/"+ID+"/activity", http.StatusSeeOther)
}

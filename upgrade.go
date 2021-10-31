package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/WeTrustPlatform/blockform/sshcmd"
	"goji.io/pat"
)

// UpgradeNode upgrades the geth to latest version
func upgradeNode(ctx context.Context, node model.Node, callback func(error)) {
	_, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		fmt.Sprintf(`
			wget --no-cache -qO- %s | bash \
			&& sudo systemctl stop geth \
			&& sudo cp geth /usr/bin/ \
			&& sudo systemctl start geth \
			`,
			cloudinit.DownloadGethSh),
	)

	if err != nil {
		log.Println(err)
		// if there are errors,
		// geth might not be running
		sshcmd.Exec(
			os.Getenv("PRIV_KEY"),
			os.Getenv("PASSPHRASE"),
			"blockform",
			node.DomainName,
			"sudo systemctl restart geth",
		)
	}

	callback(err)
}

func handleUpgrade(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, ID)
	log.Println("Upgrading geth", node.Name)
	go upgradeNode(context.Background(), node, func(err error) {
		var title = "The node has been upgraded to the latest geth"
		var eventType = model.Fine
		if err != nil {
			title = "Upgrading geth failed"
			eventType = model.Issue
		}
		db.Create(&model.Event{
			NodeID: node.ID,
			Type:   eventType,
			Title:  title,
		})
		log.Println("Done upgrading node " + node.Name)
	})
	http.Redirect(w, r, "/node/"+ID+"/activity", http.StatusSeeOther)
}

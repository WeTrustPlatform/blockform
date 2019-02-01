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

func serviceStatus(ctx context.Context, node model.Node, unit string, callback func(string, string)) {
	stdin, stderr, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		"systemctl is-active "+unit,
	)
	if err != nil {
		log.Println(err)
	}
	callback(stdin, stderr)
}

func serviceLogs(ctx context.Context, node model.Node, unit string, callback func(string, string)) {
	stdin, stderr, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		"journalctl -n 20 -q -o cat -u "+unit,
	)
	if err != nil {
		log.Println(err)
	}
	callback(stdin, stderr)
}

func handleNodeLogs(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	unit := pat.Param(r, "unit")
	if unit != "geth" && unit != "nginx" && unit != "faucet" && unit != "certbot.timer" {
		w.WriteHeader(401)
		return
	}
	node := model.Node{}
	db.Find(&node, ID)
	serviceLogs(context.Background(), node, unit, func(stdin, stderr string) {
		w.Write([]byte(stdin))
	})
}

func handleNodeStatus(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	unit := pat.Param(r, "unit")
	if unit != "geth" && unit != "nginx" && unit != "faucet" && unit != "certbot.timer" {
		w.WriteHeader(401)
		return
	}
	node := model.Node{}
	db.Find(&node, ID)
	serviceStatus(context.Background(), node, unit, func(stdin, stderr string) {
		w.Write([]byte(stdin))
	})
}

func handleNodeVersion(w http.ResponseWriter, r *http.Request) {
	ID := pat.Param(r, "id")
	node := model.Node{}
	db.Find(&node, ID)
	stdin, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		"geth version | grep ^Version",
	)
	if err != nil {
		log.Println(err)
	}
	w.Write([]byte(stdin))
}

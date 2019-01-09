package main

import (
	"net/http"

	"github.com/WeTrustPlatform/blockform/model"
)

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	var nodes []model.Node
	db.Order("created_at DESC").Find(&nodes)
	tmpl.ExecuteTemplate(w, "index.html", nodes)
}

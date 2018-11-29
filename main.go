package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/WeTrustPlatform/blockform/azure"
	"github.com/WeTrustPlatform/blockform/model"

	"github.com/alecthomas/template"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sethvargo/go-password/password"
)

// CloudProvider abstracts the behaviour of a cloud provider like AWS, Azure or
// Google Cloud. It exposes functions to create a virtual machine, install
// an ethereum node on it, and delete a virtual machine.
type CloudProvider interface {
	CreateNode(context.Context, model.Node, func())
	DeleteNode(context.Context, string, func())
}

func main() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//db.DropTableIfExists(&model.Node{})
	db.AutoMigrate(&model.Node{})

	tmpl := template.Must(template.ParseFiles(
		"templates/head.html",
		"templates/index.html",
		"templates/create.html",
		"templates/node.html",
	))

	azure := azure.NewAzure()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", basicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nodes []model.Node
		db.Find(&nodes).Order("created_at DESC")
		tmpl.ExecuteTemplate(w, "index.html", nodes)
	})))

	http.Handle("/create", basicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			tmpl.ExecuteTemplate(w, "create.html", nil)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
		}

		name := r.FormValue("name")
		if name == "" {
			name = "bf-" + password.MustGenerate(8, 4, 0, true, false)
		}
		networkID, _ := strconv.ParseInt(r.FormValue("network_id"), 10, 64)
		if networkID == 0 {
			networkID = int64(rand.Intn(1000))
		}
		APIKey := password.MustGenerate(8, 4, 0, false, false)

		node := model.Node{
			Name:        name,
			NetworkType: r.FormValue("network_type"),
			NetworkID:   uint64(networkID),
			APIKey:      APIKey,
			Status:      model.Creating,
		}

		db.Create(&node)

		go azure.CreateNode(context.Background(), node, func() {
			db.Model(&node).Update("Status", model.Deployed)
			log.Println("done creating node " + node.Name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})))

	http.Handle("/delete", basicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			w.WriteHeader(404)
		}

		db.Model(&model.Node{}).Where("name=?", name).Update("Status", model.Deleting)

		go azure.DeleteNode(context.Background(), name, func() {
			db.Where("name=?", name).Delete(&model.Node{})
			log.Println("done deleting node " + name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})))

	http.Handle("/node", basicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		node := model.Node{}
		db.Find(&node, id)
		tmpl.ExecuteTemplate(w, "node.html", node)
	})))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

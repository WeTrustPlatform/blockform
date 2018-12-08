package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/WeTrustPlatform/blockform/aws"
	"github.com/WeTrustPlatform/blockform/azure"
	"github.com/WeTrustPlatform/blockform/model"
	"goji.io/pat"

	"github.com/alecthomas/template"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sethvargo/go-password/password"
	"goji.io"
)

// CloudProvider abstracts the behaviour of a cloud provider like AWS, Azure or
// Google Cloud. It exposes functions to create a virtual machine, install
// an ethereum node on it, and delete a virtual machine.
type CloudProvider interface {
	CreateNode(context.Context, model.Node, func(string, string))
	DeleteNode(context.Context, model.Node, func())
}

var azureProvider CloudProvider
var awsProvider CloudProvider

func providerForNode(node model.Node) CloudProvider {
	var cloud CloudProvider
	switch node.CloudProvider {
	case "aws":
		cloud = awsProvider
	case "azure":
		cloud = azureProvider
	default:
		cloud = awsProvider
	}
	return cloud
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
		"templates/node_general.html",
		"templates/node_health.html",
		"templates/node_endpoints.html",
		"templates/node_actions.html",
		"templates/node_activity.html",
		"templates/node_explorer.html",
		"templates/node_sidemenu.html",
		"templates/footer.html",
	))

	azureProvider = azure.NewAzure()
	awsProvider = aws.NewAWS()

	mux := goji.NewMux()

	mux.Use(basicAuth)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle(pat.Get("/static/"), http.StripPrefix("/static/", fs))

	mux.Handle(pat.Get("/"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nodes []model.Node
		db.Find(&nodes).Order("created_at DESC")
		tmpl.ExecuteTemplate(w, "index.html", nodes)
	}))

	mux.Handle(pat.Get("/create"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "create.html", nil)
	}))

	mux.Handle(pat.Post("/create"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
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
			networkID = int64(rand.Intn(1000))
		}
		APIKey := password.MustGenerate(8, 4, 0, false, false)

		node := model.Node{
			Name:          name,
			CloudProvider: provider,
			NetworkType:   r.FormValue("network_type"),
			NetworkID:     uint64(networkID),
			APIKey:        APIKey,
			Status:        model.Creating,
		}

		db.Create(&node)

		cloud := providerForNode(node)
		go cloud.CreateNode(context.Background(), node, func(VMID, DomainName string) {
			db.Model(&node).Update("Status", model.Deployed)
			db.Model(&node).Update("VMID", VMID)
			db.Model(&node).Update("DomainName", DomainName)
			log.Println("Done creating node " + node.Name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	mux.Handle(pat.Get("/delete/:id"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID := pat.Param(r, "id")
		if ID == "" {
			w.WriteHeader(404)
		}

		node := model.Node{}
		db.Find(&node, ID)

		db.Model(&model.Node{}).Where("id=?", ID).Update("Status", model.Deleting)

		cloud := providerForNode(node)
		log.Println("Deleting node", node.Name)
		go cloud.DeleteNode(context.Background(), node, func() {
			db.Where("id=?", ID).Delete(&model.Node{})
			log.Println("Done deleting node " + node.Name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	mux.Handle(pat.Get("/node/:id"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := pat.Param(r, "id")
		http.Redirect(w, r, "/node/"+id+"/general", http.StatusSeeOther)
	}))

	mux.Handle(pat.Get("/node/:id/:tab"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := pat.Param(r, "id")
		tab := pat.Param(r, "tab")
		node := model.Node{}
		db.Find(&node, id)
		tmpl.ExecuteTemplate(w, "node_"+tab+".html", struct {
			Tab  string
			Node model.Node
		}{
			tab,
			node,
		})
	}))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

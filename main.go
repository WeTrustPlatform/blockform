package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"

	"github.com/WeTrustPlatform/blockform/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	goji "goji.io"
	"goji.io/pat"
)

// CloudProvider abstracts the behaviour of a cloud provider like AWS, Azure or
// Google Cloud. It exposes functions to create a virtual machine, install
// an ethereum node on it, and delete a virtual machine.
type CloudProvider interface {
	CreateNode(context.Context, model.Node, func(string, string), func(error))
	DeleteNode(context.Context, model.Node, func(), func(error))
}

var providers map[string]CloudProvider
var db *gorm.DB
var wg sync.WaitGroup
var tmpl *template.Template

func main() {
	var err error
	db, err = gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//db.DropTableIfExists(&model.Node{})
	//db.DropTableIfExists(&model.Event{})
	db.AutoMigrate(&model.Node{})
	db.AutoMigrate(&model.Event{})

	// terminate goroutines on sigint/sighup
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go checkSignals(sigs)

	tmpl = template.Must(template.ParseGlob("templates/*"))

	providers = makeProviders()

	mux := goji.NewMux()
	mux.Use(basicAuth)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle(pat.Get("/static/*"), http.StripPrefix("/static/", fs))

	mux.HandleFunc(pat.Get("/"), handleDashboard)
	mux.HandleFunc(pat.Get("/create"), handleCreateForm)
	mux.HandleFunc(pat.Post("/create"), handleCreate)
	mux.HandleFunc(pat.Get("/node/:id/delete"), handleNodeDelete)
	mux.HandleFunc(pat.Get("/node/:id/reboot"), handleReboot)
	mux.HandleFunc(pat.Get("/node/:id/status/:unit"), handleNodeStatus)
	mux.HandleFunc(pat.Get("/node/:id/logs/:unit"), handleNodeLogs)
	mux.HandleFunc(pat.Post("/node/:id/certbot"), handleCertbot)
	mux.HandleFunc(pat.Post("/node/:id/event/:apikey"), handleEvent)
	mux.HandleFunc(pat.Get("/node/:id"), handleNode)
	mux.HandleFunc(pat.Get("/node/:id/:tab"), handleNodeTab)
	mux.HandleFunc(pat.Get("/node/:nodeid/explorer/:class/:id"), handleNodeExplorer)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

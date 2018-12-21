package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"github.com/WeTrustPlatform/blockform/aws"
	"github.com/WeTrustPlatform/blockform/azure"
	"github.com/WeTrustPlatform/blockform/digitalocean"
	"github.com/WeTrustPlatform/blockform/gcp"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/WeTrustPlatform/blockform/sshcmd"
	"goji.io/pat"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sethvargo/go-password/password"
	goji "goji.io"
)

// CloudProvider abstracts the behaviour of a cloud provider like AWS, Azure or
// Google Cloud. It exposes functions to create a virtual machine, install
// an ethereum node on it, and delete a virtual machine.
type CloudProvider interface {
	CreateNode(context.Context, model.Node, func(string, string))
	DeleteNode(context.Context, model.Node, func(), func(error))
}

var providers map[string]CloudProvider

// RebootNode reboots the VM where the node is hosted
func rebootNode(ctx context.Context, node model.Node, callback func()) {
	err := sshcmd.Exec(
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

func makeProviders() map[string]CloudProvider {
	prov := make(map[string]CloudProvider)
	azureProvider, err := azure.NewAzure()
	if err == nil {
		prov["azure"] = azureProvider
	}
	awsProvider, err := aws.NewAWS()
	if err == nil {
		prov["aws"] = awsProvider
	}
	doProvider, err := digitalocean.NewDigitalOcean()
	if err == nil {
		prov["digitalocean"] = doProvider
	}
	gcpProvider, err := gcp.NewGCP()
	if err == nil {
		prov["gcp"] = gcpProvider
	}
	if len(prov) == 0 {
		log.Println("No cloud provider, you won't be able to create nodes")
	}
	return prov
}

func main() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//db.DropTableIfExists(&model.Node{})
	db.AutoMigrate(&model.Node{})

	tmpl := template.Must(template.ParseGlob("templates/*"))

	providers = makeProviders()

	mux := goji.NewMux()

	mux.Use(basicAuth)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle(pat.Get("/static/*"), http.StripPrefix("/static/", fs))

	mux.HandleFunc(pat.Get("/"), func(w http.ResponseWriter, r *http.Request) {
		var nodes []model.Node
		db.Order("created_at DESC").Find(&nodes)
		tmpl.ExecuteTemplate(w, "index.html", nodes)
	})

	mux.HandleFunc(pat.Get("/create"), func(w http.ResponseWriter, r *http.Request) {
		keys := reflect.ValueOf(providers).MapKeys()
		tmpl.ExecuteTemplate(w, "create.html", keys)
	})

	mux.HandleFunc(pat.Post("/create"), func(w http.ResponseWriter, r *http.Request) {
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
			rand.Seed(time.Now().UnixNano())
			networkID = int64(rand.Intn(65536))
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

		cloud := providers[node.CloudProvider]
		go cloud.CreateNode(context.Background(), node, func(VMID, DomainName string) {
			db.Model(&node).Update("Status", model.Deployed)
			db.Model(&node).Update("VMID", VMID)
			db.Model(&node).Update("DomainName", DomainName)
			log.Println("Done creating node " + node.Name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc(pat.Get("/delete/:id"), func(w http.ResponseWriter, r *http.Request) {
		ID := pat.Param(r, "id")
		node := model.Node{}
		db.Find(&node, ID)

		db.Model(&model.Node{}).Where("id=?", ID).Update("Status", model.Deleting)

		cloud := providers[node.CloudProvider]
		log.Println("Deleting node", node.Name)
		go cloud.DeleteNode(context.Background(), node,
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

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc(pat.Get("/reboot/:id"), func(w http.ResponseWriter, r *http.Request) {
		ID := pat.Param(r, "id")
		node := model.Node{}
		db.Find(&node, ID)
		log.Println("Rebooting node", node.Name)
		go rebootNode(context.Background(), node, func() {
			log.Println("Done rebooting node " + node.Name)
		})
		http.Redirect(w, r, "/node/"+ID+"/actions", http.StatusSeeOther)
	})

	mux.HandleFunc(pat.Get("/node/:id"), func(w http.ResponseWriter, r *http.Request) {
		id := pat.Param(r, "id")
		http.Redirect(w, r, "/node/"+id+"/general", http.StatusSeeOther)
	})

	mux.HandleFunc(pat.Get("/node/:id/:tab"), func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc(pat.Get("/node/:nodeid/explorer/:class/:id"), func(w http.ResponseWriter, r *http.Request) {
		nodeID := pat.Param(r, "nodeid")
		tab := "explorer"
		class := pat.Param(r, "class")
		id := pat.Param(r, "id")
		node := model.Node{}
		db.Find(&node, nodeID)
		tmpl.ExecuteTemplate(w, "node_"+tab+"_"+class+".html", struct {
			Tab  string
			ID   string
			Node model.Node
		}{
			tab,
			id,
			node,
		})
	})

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}

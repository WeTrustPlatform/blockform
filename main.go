package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/alecthomas/template"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sethvargo/go-password/password"
)

func main() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//db.DropTableIfExists(&model.Node{})
	db.AutoMigrate(&model.Node{})

	tmpl := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/create.html",
	))

	authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to get Azure OAuth config: %v\n", err)
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nodes []model.Node
		db.Find(&nodes).Order("created_at DESC")
		tmpl.ExecuteTemplate(w, "index.html", nodes)
	}))

	http.Handle("/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "create.html", nil)
	}))

	http.Handle("/create/public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
		}

		name := r.FormValue("name")
		if name == "" {
			name = "bf-" + password.MustGenerate(8, 4, 0, true, false)
		}
		networkID, _ := strconv.ParseInt(r.FormValue("network_id"), 10, 64)
		APIKey := password.MustGenerate(8, 4, 0, false, false)

		node := model.Node{
			Name:        name,
			NetworkType: model.Public,
			NetworkID:   uint64(networkID),
			APIKey:      APIKey,
			Status:      model.Creating,
		}

		db.Create(&node)

		go createNode(context.Background(), node, func() {
			db.Model(&node).Update("Status", model.Deployed)
			log.Println("done creating node " + node.Name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	http.Handle("/create/private", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not implemented yet"))
	}))

	http.Handle("/delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			w.WriteHeader(404)
		}

		db.Model(&model.Node{}).Where("name=?", name).Update("Status", model.Deleting)

		go deleteNode(context.Background(), name, func() {
			db.Where("name=?", name).Delete(&model.Node{})
			log.Println("done deleting node " + name)
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

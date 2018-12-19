package gcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"

	"github.com/WeTrustPlatform/blockform/model"
	"google.golang.org/api/compute/v1"
)

// GCP is an implementation of CloudProvider for Google Cloud Platform
type GCP struct {
	service *compute.Service
}

// NewGCP instantiates a new GCP CloudProvider
func NewGCP() GCP {
	var gc GCP

	ctx := context.Background()
	data, err := ioutil.ReadFile("/Users/kivutar/Downloads/file (3).txt")
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(data,
		"https://www.googleapis.com/auth/compute",
	)
	if err != nil {
		log.Fatal(err)
	}

	gc.service, _ = compute.New(conf.Client(ctx))
	return gc
}

// CreateNode creates a google compute engine instance
func (gc GCP) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {
	log.Println("Creating a node on GCP")

	project, err := gc.service.Projects.Get("blockform").Do()
	if err != nil {
		log.Println(err)
	}

	zones, _ := gc.service.Zones.List(project.Name).Do()

	fmt.Println(zones)

	gc.service.Instances.Insert(project.Name, "us-west2-a", &compute.Instance{
		Name: node.Name,
	}).Do()

	callback(node.Name, node.Name)
}

// DeleteNode deletes the google compute engine instance
func (gc GCP) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	log.Println("Deleting a node on GCP")
	onSuccess()
}

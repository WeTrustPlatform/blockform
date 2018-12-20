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

	prefix := "https://www.googleapis.com/compute/v1/projects/" + project.Name

	op, err := gc.service.Instances.Insert(project.Name, "us-west2-a", &compute.Instance{
		Name:        node.Name,
		MachineType: prefix + "/zones/us-west2-a/machineTypes/g1-small",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    node.Name,
					DiskSizeGb:  10,
					SourceImage: "projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20181203a",
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: prefix + "/global/networks/default",
			},
		},
		Metadata: &compute.Metadata{},
		Tags: &compute.Tags{
			Items: []string{"blockform"},
		},
	}).Do()
	if err != nil {
		log.Println(err)
	}

	fmt.Println(op)

	callback(node.Name, "")
}

// DeleteNode deletes the google compute engine instance
func (gc GCP) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	log.Println("Deleting a node on GCP")

	_, err := gc.service.Instances.Delete("blockform", "us-west2-a", node.VMID).Do()
	if err != nil {
		onError(err)
		return
	}

	onSuccess()
}

package gcp

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/WeTrustPlatform/blockform/cloudinit"

	"golang.org/x/oauth2/google"

	"github.com/WeTrustPlatform/blockform/model"
	"google.golang.org/api/compute/v1"
)

// GCP is an implementation of CloudProvider for Google Cloud Platform
type GCP struct {
	service *compute.Service
}

// NewGCP instantiates a new GCP CloudProvider
func NewGCP() (*GCP, error) {
	if os.Getenv("GCP_JSON") == "" ||
		os.Getenv("GCP_PROJECT") == "" {
		err := errors.New("GCP_JSON or GCP_PROJECT is not set")
		log.Println("Cloudn't create GCP:", err)
		return nil, err
	}

	conf, err := google.JWTConfigFromJSON(
		[]byte(os.Getenv("GCP_JSON")),
		"https://www.googleapis.com/auth/compute",
	)
	if err != nil {
		log.Fatal(err)
	}

	var gc GCP
	gc.service, _ = compute.New(conf.Client(context.Background()))
	return &gc, nil
}

var (
	zone    = "asia-east1-a"
	project = os.Getenv("GCP_PROJECT")
	prefix  = "https://www.googleapis.com/compute/v1/projects/" + project
	size    = "g1-small"
	image   = "projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20181203a"
)

// CreateNode creates a google compute engine instance
func (gc GCP) CreateNode(ctx context.Context, node model.Node, callback func(string, string), onError func(error)) {
	log.Println("Creating a node on GCP")

	customData := cloudinit.CustomData(node, "/dev/sdb")

	sizeForMode := map[string]int64{
		model.Full:  2000,
		model.Fast:  200,
		model.Light: 20,
	}

	insertOP, err := gc.service.Instances.Insert(project, zone, &compute.Instance{
		Name:        node.Name,
		MachineType: prefix + "/zones/" + zone + "/machineTypes/" + size,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    node.Name + "-os",
					DiskSizeGb:  30,
					SourceImage: image,
				},
			},
			{
				AutoDelete: true,
				Boot:       false,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:   node.Name + "-data",
					DiskType:   prefix + "/zones/" + zone + "/diskTypes/pd-ssd",
					DiskSizeGb: sizeForMode[node.SyncMode],
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
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "user-data",
					Value: &customData,
				},
			},
		},
		Tags: &compute.Tags{Items: []string{project, node.Name}},
	}).Do()

	if err != nil {
		onError(err)
		log.Println(err)
		return
	}

	for {
		time.Sleep(10 * time.Second)
		op, err := gc.service.ZoneOperations.Get(project, zone, insertOP.Name).Do()
		if err != nil {
			log.Println(err)
			break
		}
		if op.Status == "DONE" {
			break
		}
	}

	vm, err := gc.service.Instances.Get(project, zone, node.Name).Do()
	if err != nil {
		onError(err)
		log.Println(err)
		return
	}

	net := vm.NetworkInterfaces[0].Network

	_, err = gc.service.Firewalls.Insert(project, &compute.Firewall{
		Name:         node.Name,
		TargetTags:   []string{node.Name},
		Network:      net,
		SourceRanges: []string{"0.0.0.0/0"},
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "TCP",
				Ports:      []string{"22", "8545", "8546", "8080", "80"},
			},
			{
				IPProtocol: "UDP",
				Ports:      []string{"30303"},
			},
		},
	}).Do()
	if err != nil {
		onError(err)
		log.Println(err)
		return
	}

	callback(node.Name, vm.NetworkInterfaces[0].AccessConfigs[0].NatIP)
}

// DeleteNode deletes the google compute engine instance
func (gc GCP) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	log.Println("Deleting a node on GCP")

	_, err := gc.service.Instances.Delete(project, zone, node.VMID).Do()
	if err != nil {
		onError(err)
		return
	}

	onSuccess()
}

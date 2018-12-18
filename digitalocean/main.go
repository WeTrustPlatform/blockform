package digitalocean

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"

	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
)

// DigitalOcean is an implementation of CloudProvider for DigitalOcean
type DigitalOcean struct {
	client *godo.Client
}

// TokenSource stores the OAuth2 access token string
type TokenSource struct {
	AccessToken string
}

// Token returns the oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// NewDigitalOcean instantiates a new DigitalOcean CloudProvider
func NewDigitalOcean() DigitalOcean {
	var do DigitalOcean
	tokenSource := &TokenSource{
		AccessToken: os.Getenv("DO_TOKEN"),
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	do.client = godo.NewClient(oauthClient)
	return do
}

// CreateNode creates a volume and a droplet and installs geth.
func (do DigitalOcean) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {

	customData := cloudinit.CustomData(node, "/dev/sda")

	vol, _, err := do.client.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
		Name:          node.Name,
		Region:        "sfo2",
		SizeGigaBytes: 200,
	})
	if err != nil {
		log.Println(err)
	}

	newDroplet, _, err := do.client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:   node.Name,
		Region: "sfo2",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-18-04-x64",
		},
		IPv6: true,
		Tags: []string{"blockform", node.Name},
		Volumes: []godo.DropletCreateVolume{
			{
				ID: vol.ID,
			},
		},
		UserData: customData,
	})
	if err != nil {
		log.Println(err)
	}

	_, _, err = do.client.Firewalls.Create(ctx, &godo.FirewallRequest{
		Name:       node.Name,
		DropletIDs: []int{newDroplet.ID},
		InboundRules: []godo.InboundRule{
			{Protocol: "TCP", PortRange: "22", Sources: &godo.Sources{}},
			{Protocol: "TCP", PortRange: "80", Sources: &godo.Sources{}},
			{Protocol: "TCP", PortRange: "8080", Sources: &godo.Sources{}},
			{Protocol: "TCP", PortRange: "8545", Sources: &godo.Sources{}},
			{Protocol: "TCP", PortRange: "8546", Sources: &godo.Sources{}},
			{Protocol: "UDP", PortRange: "30303", Sources: &godo.Sources{}},
		},
	})
	if err != nil {
		log.Println(err)
	}

	time.Sleep(40 * time.Second)

	droplet, _, _ := do.client.Droplets.Get(ctx, newDroplet.ID)

	ipv4, _ := droplet.PublicIPv4()
	callback(fmt.Sprintf("%d", droplet.ID), ipv4)
}

// DeleteNode deletes the droplet and the attached volume and firewall.
// The firewall is deleted first, because the droplet needs to exists when
// looking for the firewall. The volume is deleted last, because we can't delete
// a volume attached to an existing droplet.
func (do DigitalOcean) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	id, err := strconv.ParseInt(node.VMID, 10, 64)
	if err != nil {
		onError(err)
		return
	}

	droplet, _, err := do.client.Droplets.Get(ctx, int(id))
	if err != nil {
		onError(err)
		return
	}

	fws, _, err := do.client.Firewalls.ListByDroplet(ctx, droplet.ID, &godo.ListOptions{
		Page:    0,
		PerPage: 10,
	})
	if err != nil {
		onError(err)
		return
	}

	_, err = do.client.Firewalls.Delete(ctx, fws[0].ID)
	if err != nil {
		onError(err)
		return
	}

	_, err = do.client.Droplets.DeleteByTag(ctx, node.Name)
	if err != nil {
		onError(err)
		return
	}

	time.Sleep(20 * time.Second)

	_, err = do.client.Storage.DeleteVolume(ctx, droplet.VolumeIDs[0])
	if err != nil {
		onError(err)
		return
	}

	onSuccess()
}

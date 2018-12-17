package digitalocean

import (
	"context"
	"fmt"
	"os"
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

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func NewDigitalOcean() DigitalOcean {
	var do DigitalOcean
	tokenSource := &TokenSource{
		AccessToken: os.Getenv("DO_TOKEN"),
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	do.client = godo.NewClient(oauthClient)
	return do
}

func (do DigitalOcean) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {

	customData := cloudinit.CustomData(node, "/dev/sda")

	vol, _, _ := do.client.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
		Name:          node.Name + "-volume",
		Region:        "sfo2",
		SizeGigaBytes: 10,
	})

	time.Sleep(30 * time.Second)

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
		fmt.Printf("Something bad happened: %s\n\n", err)
	}

	time.Sleep(40 * time.Second)

	droplet, _, _ := do.client.Droplets.Get(ctx, newDroplet.ID)

	ipv4, _ := droplet.PublicIPv4()
	callback(droplet.URN(), ipv4)
}

func (do DigitalOcean) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	_, err := do.client.Droplets.DeleteByTag(ctx, node.Name)
	if err != nil {
		onError(err)
		return
	}
	onSuccess()
}

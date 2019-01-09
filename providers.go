package main

import (
	"log"

	"github.com/WeTrustPlatform/blockform/aws"
	"github.com/WeTrustPlatform/blockform/azure"
	"github.com/WeTrustPlatform/blockform/digitalocean"
	"github.com/WeTrustPlatform/blockform/gcp"
)

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

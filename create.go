package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/WeTrustPlatform/blockform/model"
)

var (
	authorizer     autorest.Authorizer
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	location       = "westus2"
)

func createGroup(ctx context.Context, groupName string) (resources.Group, error) {
	groupsClient := resources.NewGroupsClient(subscriptionID)
	groupsClient.Authorizer = authorizer

	log.Printf("creating resource group '%s' on location: %v\n", groupName, location)
	return groupsClient.CreateOrUpdate(
		ctx,
		groupName,
		resources.Group{
			Location: to.StringPtr(location),
		})
}

func getCustomData(node model.Node) string {
	var data []byte
	switch node.NetworkID {
	case 1:
		data, _ = ioutil.ReadFile("cloud-init/mainnet.yml")
	case 4:
		data, _ = ioutil.ReadFile("cloud-init/rinkeby.yml")
	}

	if node.NetworkType == model.Private {
		data, _ = ioutil.ReadFile("cloud-init/private.yml")
	}

	str := string(data)
	str = strings.Replace(str, "@@API_KEY@@", node.APIKey, -1)
	str = strings.Replace(str, "@@NET_ID@@", fmt.Sprintf("%d", node.NetworkID), -1)

	return base64.StdEncoding.EncodeToString([]byte(str))
}

func createNode(ctx context.Context, node model.Node, callback func()) {
	group, err := createGroup(ctx, node.Name)
	if err != nil {
		log.Printf("cannot create group: %v\n", err)
	}

	vmClient := compute.NewVirtualMachinesClient(subscriptionID)
	vmClient.Authorizer = authorizer

	deploymentsClient := resources.NewDeploymentsClient(subscriptionID)
	deploymentsClient.Authorizer = authorizer

	template, err := readJSON("vm-templates/small.json")
	if err != nil {
		log.Println(err)
	}

	customData := getCustomData(node)

	params := map[string]interface{}{
		"vm_user":     map[string]interface{}{"value": "wetrust"},
		"vm_password": map[string]interface{}{"value": "wetrustwetrustO*"},
		"dns_prefix":  map[string]interface{}{"value": *group.Name},
		"custom_data": map[string]interface{}{"value": customData},
	}

	deploymentFuture, err := deploymentsClient.CreateOrUpdate(
		ctx,
		*group.Name,
		*group.Name+"DEP",
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: &params,
				Mode:       resources.Incremental,
			},
		},
	)
	if err != nil {
		log.Println(err)
	}
	err = deploymentFuture.Future.WaitForCompletionRef(ctx, deploymentsClient.BaseClient.Client)
	if err != nil {
		log.Println(err)
	}

	callback()
}

func readJSON(path string) (*map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	contents := make(map[string]interface{})
	json.Unmarshal(data, &contents)
	return &contents, nil
}

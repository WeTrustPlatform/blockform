package azure

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
)

// Azure is an implementation of CloudProvider for Microsoft Azure
type Azure struct {
	groupsClient      resources.GroupsClient
	deploymentsClient resources.DeploymentsClient
	authorizer        autorest.Authorizer
}

var (
	location = "westus2"
)

// NewAzure instanciates an Azure CloudProvider and sets important members
// like the authorizer.
func NewAzure() Azure {
	var az Azure
	var err error
	az.authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to get Azure OAuth config: %v\n", err)
	}
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	az.groupsClient = resources.NewGroupsClient(subscriptionID)
	az.groupsClient.Authorizer = az.authorizer
	az.deploymentsClient = resources.NewDeploymentsClient(subscriptionID)
	az.deploymentsClient.Authorizer = az.authorizer
	return az
}

// CreateNode will create an azure VM and install a geth node using cloud-init
// and execute the callback when done.
func (az Azure) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {
	group, err := az.createGroup(ctx, node.Name)
	if err != nil {
		log.Printf("cannot create group: %v\n", err)
	}

	template, err := readJSON("azure/small.json")
	if err != nil {
		log.Println(err)
	}

	customData := cloudinit.CustomData(node, "/dev/sdc")

	params := map[string]interface{}{
		"vm_user":     map[string]interface{}{"value": "blockform"},
		"pub_key":     map[string]interface{}{"value": os.Getenv("PUB_KEY")},
		"dns_prefix":  map[string]interface{}{"value": *group.Name},
		"custom_data": map[string]interface{}{"value": customData},
	}

	deploymentFuture, err := az.deploymentsClient.CreateOrUpdate(
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

	err = deploymentFuture.Future.WaitForCompletionRef(ctx, az.deploymentsClient.BaseClient.Client)
	if err != nil {
		log.Println(err)
	}

	domainName := node.Name + ".westus2.cloudapp.azure.com"
	callback(*group.Name, domainName)
}

// DeleteNode deletes the resource group with all the resources in it and
// executes the callback when it's done.
func (az Azure) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	groupsDeleteFuture, err := az.groupsClient.Delete(ctx, node.VMID)
	if err != nil {
		onError(err)
		return
	}

	err = groupsDeleteFuture.Future.WaitForCompletionRef(ctx, az.groupsClient.BaseClient.Client)
	if err != nil {
		onError(err)
		return
	}

	onSuccess()
}

func (az Azure) createGroup(ctx context.Context, groupName string) (resources.Group, error) {
	log.Printf("creating resource group '%s' on location: %v\n", groupName, location)
	return az.groupsClient.CreateOrUpdate(
		ctx,
		groupName,
		resources.Group{
			Location: to.StringPtr(location),
		})
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

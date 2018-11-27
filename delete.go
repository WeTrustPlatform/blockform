package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
)

func deleteNode(ctx context.Context, name string, callback func()) {
	log.Println("deleting resource group " + name)
	groupsClient := resources.NewGroupsClient(subscriptionID)
	groupsClient.Authorizer = authorizer
	groupsDeleteFuture, err := groupsClient.Delete(ctx, name)

	if err != nil {
		log.Println(err)
	}
	err = groupsDeleteFuture.Future.WaitForCompletionRef(ctx, groupsClient.BaseClient.Client)
	if err != nil {
		log.Println(err)
	}

	callback()
}

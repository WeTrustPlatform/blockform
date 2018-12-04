package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AWS is an implementation of CloudProvider for Amazon Web Services
type AWS struct {
	svc *ec2.EC2
}

// NewAWS instanciates an AWS CloudProvider and creates an EC2 session.
func NewAWS() AWS {
	var aw AWS

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))

	sess.Handlers.Send.PushFront(func(r *request.Request) {
		// Log every request made and its payload
		log.Printf("Request: %v/%v, Payload: %v\n",
			r.ClientInfo.ServiceName, r.Operation, r.Params)
	})

	aw.svc = ec2.New(sess)

	return aw
}

// CreateNode created an EC2 instance and setups geth
func (aw AWS) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {

	importResult, err := aw.svc.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String("blockform"),
		PublicKeyMaterial: []byte(os.Getenv("PUB_KEY")),
	})
	fmt.Println(importResult, err)

	customData := cloudinit.CustomData(node)

	run, err := aw.svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0f9cf087c1f27d9b1"), // Ubuntu 16.04
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      importResult.KeyName,
		UserData:     aws.String(customData),
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
	}

	VMID := *run.Instances[0].InstanceId

	fmt.Println("Created instance", VMID)

	_, errtag := aw.svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(VMID)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(node.Name),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", VMID, errtag)
	}

	for {
		time.Sleep(30 * time.Second)

		status, _ := aw.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			InstanceIds: []*string{aws.String(VMID)},
		})

		fmt.Println(status)

		if len(status.InstanceStatuses) > 0 {
			if *status.InstanceStatuses[0].SystemStatus.Status == "ok" {
				break
			}
		}
	}

	instances, _ := aw.svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(VMID)},
	})

	publicDNSName := *instances.Reservations[0].Instances[0].PublicDnsName

	callback(VMID, publicDNSName)
}

// DeleteNode will delete the AWS node
func (aw AWS) DeleteNode(ctx context.Context, VMID string, callback func()) {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(VMID)},
	}

	result, err := aw.svc.TerminateInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)

	callback()
}

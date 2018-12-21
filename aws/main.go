package aws

import (
	"context"
	"log"
	"time"

	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/aws/aws-sdk-go/aws"
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
func NewAWS() (*AWS, error) {
	var aw AWS

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})
	if err != nil {
		log.Println("Failed to create AWS session:", err)
		return nil, err
	}

	// Log every request made and its payload
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		log.Printf("Request: %v/%v, Payload: %v\n",
			r.ClientInfo.ServiceName, r.Operation, r.Params)
	})

	aw.svc = ec2.New(sess)

	return &aw, nil
}

// CreateNode created an EC2 instance and setups geth
func (aw AWS) CreateNode(ctx context.Context, node model.Node, callback func(string, string)) {

	sgName := node.Name // name the security group after the node name
	aw.createSecurityGroup(sgName)

	customData := cloudinit.EncodedCustomData(node, "/dev/xvdc")

	run, err := aw.svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:        aws.String("ami-0d2505740b82f7948"), // Ubuntu 18.04
		InstanceType:   aws.String("t2.medium"),
		MinCount:       aws.Int64(1),
		MaxCount:       aws.Int64(1),
		SecurityGroups: []*string{aws.String(sgName)},
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sdc"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize:          aws.Int64(200),
					VolumeType:          aws.String("gp2"),
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		UserData: aws.String(customData),
	})
	if err != nil {
		log.Println("Could not create instance", err)
	}

	VMID := *run.Instances[0].InstanceId

	log.Println("Created instance", VMID)

	_, err = aw.svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(VMID)},
		Tags: []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String(node.Name)},
			{Key: aws.String("Creator"), Value: aws.String("blockform")},
		},
	})
	if err != nil {
		log.Println("Could not create tags for instance", VMID, err)
	}

	// Wait until the instance is fully deployed
	for {
		time.Sleep(30 * time.Second)

		status, _ := aw.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			InstanceIds: []*string{aws.String(VMID)},
		})

		log.Println(status)

		if len(status.InstanceStatuses) > 0 {
			if *status.InstanceStatuses[0].SystemStatus.Status == "ok" {
				break
			}
		}
	}

	instances, err := aw.svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(VMID)},
	})
	if err != nil {
		log.Println("Could not describe instance", VMID, err)
	}

	publicDNSName := *instances.Reservations[0].Instances[0].PublicDnsName

	callback(VMID, publicDNSName)
}

// DeleteNode will delete the AWS node
func (aw AWS) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	termInst, err := aw.svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(node.VMID)},
	})
	if err != nil {
		log.Println("Could not delete instance:", err)
		onError(err)
		return
	}
	log.Println(termInst)

	// Wait until the instance is fully terminated
	for {
		time.Sleep(30 * time.Second)

		status, _ := aw.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
			InstanceIds: []*string{aws.String(node.VMID)},
		})

		log.Println(status)

		if len(status.InstanceStatuses) == 0 {
			break
		}
	}

	delSG, err := aw.svc.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupName: aws.String(node.Name),
	})
	if err != nil {
		log.Println("Could not delete security group:", err)
		onError(err)
		return
	}
	log.Println(delSG)

	onSuccess()
}

// createSecurityGroup creates the security group with the VPC, name and description.
func (aw AWS) createSecurityGroup(name string) {
	sg, err := aw.svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String("Security group for blockform"),
	})
	if err != nil {
		log.Println("Unable to create security group:", name, err)
	}
	log.Println("Created security group", aws.StringValue(sg.GroupId))

	_, err = aw.svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupName: aws.String(name),
		IpPermissions: []*ec2.IpPermission{
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(22).
				SetToPort(22).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(8545).
				SetToPort(8545).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(8546).
				SetToPort(8546).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(8080).
				SetToPort(8080).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("tcp").
				SetFromPort(30303).
				SetToPort(30303).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
			(&ec2.IpPermission{}).
				SetIpProtocol("udp").
				SetFromPort(30303).
				SetToPort(30303).
				SetIpRanges([]*ec2.IpRange{
					(&ec2.IpRange{}).SetCidrIp("0.0.0.0/0"),
				}).
				SetIpv6Ranges([]*ec2.Ipv6Range{
					(&ec2.Ipv6Range{}).SetCidrIpv6("::/0"),
				}),
		},
	})
	if err != nil {
		log.Println("Unable to set security group ingress:", name, err)
	}
}

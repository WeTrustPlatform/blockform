# Instructions:
# 1. Create the terraform.tfvars file to store secrets
# access_key = "<your aws key. Must have permissions to create VPC and EC2>"
# access_secret = "<your aws secret>"
# ssh_public_key_path = "Default value: ~/.ssh/id_rsa.pub"
# ssh_private_key_path = "Default value: ~/.ssh/id_rsa"
#
# 2. Run: terraform init
# 3. Run: terraform apply
# 4. Review the resources to be created
# 5. Enter yes and wait for magic to happen!

# TODO
# Set up custom domain and SSL

# Total 2 of steps:
# Step 1 (aws.tf): Provision AWS with attached storage
# Step 2 (geth.tf): Set up all the necessary for geth

variable "access_key" {}

variable "access_secret" {}

variable "region" {
  default = "us-west-2"
}

# This ami_id depends on region
# Ubuntu 18.04 hvm:ebs-ssd
variable "ami_id" {
  default = "ami-079b4e9085609225c"
}

variable "env" {
  default = "mainnet"
}

variable "app_name" {
  default = "geth"
}

variable "nginx_conf" {
  default = "nginx.conf"
}

variable "ssh_public_key_path" {
  default = "~/.ssh/id_rsa.pub"
}

variable "ssh_private_key_path" {
  default = "~/.ssh/id_rsa"
}

locals {
  common_tags = {
    Terraform   = "true"
    Environment = "${var.env}"
    App         = "${var.app_name}"
  }
}

resource "random_string" "username" {
  length  = 8
  special = false
}

resource "random_string" "password" {
  length  = 24
  special = false
}

provider "aws" {
  access_key = "${var.access_key}"
  secret_key = "${var.access_secret}"
  region     = "${var.region}"
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  name   = "${var.app_name}"

  cidr           = "172.31.0.0/16"
  public_subnets = ["172.31.32.0/20", "172.31.64.0/20"]

  enable_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true

  azs  = ["${var.region}a", "${var.region}b"]
  tags = "${local.common_tags}"
}

module "geth_instance_sg" {
  source = "terraform-aws-modules/security-group/aws"
  name   = "geth-instance-sg"

  description = "Security group for geth-instance"
  vpc_id      = "${module.vpc.vpc_id}"

  # Allow all out going ports
  egress_cidr_blocks      = ["0.0.0.0/0"]
  egress_ipv6_cidr_blocks = ["::/0"]
  egress_rules            = ["all-tcp", "all-udp"]

  # TODO whitelisted IP for ssh in
  ingress_cidr_blocks = ["0.0.0.0/0"]
  ingress_rules       = ["ssh-tcp"]

  # Allow 8545 and 30303 incoming
  ingress_with_cidr_blocks = [
    {
      from_port        = 8545
      to_port          = 8545
      protocol         = "tcp"
      description      = "Geth RPC"
      cidr_blocks      = "0.0.0.0/0"
      ipv6_cidr_blocks = "::/0"
    },
    {
      from_port        = 30303
      to_port          = 30303
      protocol         = "tcp"
      description      = "Geth TCP Listener"
      cidr_blocks      = "0.0.0.0/0"
      ipv6_cidr_blocks = "::/0"
    },
    {
      from_port        = 30303
      to_port          = 30303
      protocol         = "udp"
      description      = "Geth UDP Discovery"
      cidr_blocks      = "0.0.0.0/0"
      ipv6_cidr_blocks = "::/0"
    },
  ]

  tags = "${local.common_tags}"
}

resource "aws_key_pair" "deployer" {
  key_name   = "${var.app_name}-${var.env}"
  public_key = "${file(var.ssh_public_key_path)}"
}

resource "aws_instance" "geth_instance" {
  ami                         = "${var.ami_id}"
  associate_public_ip_address = true
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.deployer.key_name}"
  monitoring                  = true
  vpc_security_group_ids      = ["${module.geth_instance_sg.this_security_group_id}"]
  subnet_id                   = "${module.vpc.public_subnets[0]}"

  ebs_block_device = {
    device_name           = "/dev/sdg"
    delete_on_termination = false
    volume_size           = 500
    volume_type           = "gp2"
  }

  tags = "${local.common_tags}"
}

output "aws_instance_public_dns" {
  value = "${aws_instance.geth_instance.public_dns}"
}

output "rpc_endpoint" {
  value = "http://${random_string.username.result}:${random_string.password.result}@${aws_instance.geth_instance.public_dns}:8545"
}

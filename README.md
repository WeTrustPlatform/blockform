# BlockForm - Ethereum Blockchain As A Service

[![GoDoc](https://godoc.org/github.com/WeTrustPlatform/blockform?status.svg)](https://godoc.org/github.com/WeTrustPlatform/blockform)
[![Go Report Card](https://goreportcard.com/badge/github.com/WeTrustPlatform/blockform)](https://goreportcard.com/report/github.com/WeTrustPlatform/blockform)
[![Build Status](https://travis-ci.org/WeTrustPlatform/blockform.svg?branch=master)](https://travis-ci.org/WeTrustPlatform/blockform)

BlockForm helps you deploy Ethereum nodes to the cloud and manage them.
It takes care of creating a virtual machine in one of the supported cloud
providers (currently Azure, AWS and DigitalOcean) and will install geth in the
best possible way.

Blockform is a free software. It is tailored for teams or individuals whose
work depends on a fleet of Ethereum nodes. You have to install your own
instance. So the installation process has been made as simple as possible.

![Screenshot](https://github.com/WeTrustPlatform/blockform/raw/master/static/screenshot.png)

### Features

 - Deploy geth nodes for public networks
 - Create your own private Ethereum network with Proof of Authority
 - Support multiple cloud providers: Azure, AWS and DigitalOcean
 - RPC endpoints protected by an API key
 - WebSocket RPC endpoints
 - Basic faucet web service for your private network
 - Node data saved to an SSD
 - Basic heathchecks
 - Basic authentication for the UI
 - Conpatible with MetaMask, geth attach, and web3.js
 - Easy to deploy on Heroku

### Usage

For Azure, generate a service principal:

    az login
    az ad sp create-for-rbac --sdk-auth

For AWS, go to IAM Management console, and generate an Access Key for your
user in the Security Cretendials tab.

For DigitalOcean, go to API and generate a Personal access token.

For Google Cloud Platform:

 - Create a project, for example `my-blockform`. If the name is not long enough,
 google will append some random string that you will need to export in the next
 steps.
 - In *API and Services*, enable the *Compute Engine API* for this project.
 - In *API and Services*, go to the *Credentials* tab and generate an *Account
 Service Key* as a JSON file for the *Compute Engine default service account*.

Set environment variables:

    export AZURE_TENANT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_SECRET="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_SUBSCRIPTION_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AWS_ACCESS_KEY_ID="xxxxxxxxxx"
    export AWS_SECRET_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxx"
    export DO_TOKEN="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    export GCP_JSON="{xxxx}"
    export GCP_PROJECT="my-blockform"
    export PORT=3000
    export DATABASE_URL="postgres://user:pass@localhost:5432/blockform?sslmode=disable"
    export BASIC_AUTH_USER=xxxxx
    export BASIC_AUTH_PASS=xxxxx
    export PUB_KEY="xxxxx" # your SSH public key to login the VMs

Get, build and run:

    go get github.com/WeTrustPlatform/blockform
    cd $GOPATH/src/github.com/WeTrustPlatform/blockform
    go build && ./blockform

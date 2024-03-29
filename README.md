# BlockForm - Ethereum Blockchain As A Service

[![GoDoc](https://godoc.org/github.com/WeTrustPlatform/blockform?status.svg)](https://godoc.org/github.com/WeTrustPlatform/blockform)
[![Go Report Card](https://goreportcard.com/badge/github.com/WeTrustPlatform/blockform)](https://goreportcard.com/report/github.com/WeTrustPlatform/blockform)
[![Build Status](https://travis-ci.org/WeTrustPlatform/blockform.svg?branch=master)](https://travis-ci.org/WeTrustPlatform/blockform)

BlockForm helps you deploy Ethereum nodes to the cloud and manage them.
It takes care of creating a virtual machine in one of the supported cloud
providers (currently Azure, AWS, GCP and DigitalOcean) and will install geth in
the best possible way.

Blockform is a free software. It is tailored for teams or individuals whose
work depends on a fleet of Ethereum nodes. You have to install your own
instance. So the installation process has been made as simple as possible.

![Screenshot](https://github.com/WeTrustPlatform/blockform/raw/master/static/screenshot.png)

### Features

 - Support public networks: Mainnet, Rinkeby, Ropsten
 - Allows you to create your own private ethereum network with Proof of Authority
 - Support multiple cloud providers: AWS, GCP and DigitalOcean
 - Support multiple synchronization modes: Full, Fast and Light
 - Setup a firewall for you
 - SSL certificates through certbot
 - HTTP and WebSocket RPC endpoints protected by an API key
 - Basic faucet web service for your private network
 - Node data is saved to an SSD for faster sync
 - Basic heathchecks and logs display
 - Basic authentication for the UI
 - Tiny chain explorer to browse your blocks, transactions and accounts
 - Compatible with MetaMask, geth attach, and web3.js
 - Easy deployments on Heroku
 - Optional Slack integration for alerts

Note: Azure is broken and removed after the update to go 1.17 commit - this also
help resolve the security issue with azure's dep github.com/dgrijalva/jwt-go

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

Set environment variables in `.env`:

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
    export APP_ENV=development | production
    export DATABASE_URL="postgres://user:pass@localhost:5432/blockform?sslmode=disable"
    export BASIC_AUTH_USER=xxxxx
    export BASIC_AUTH_PASS=xxxxx
    export SITE_URL="http://localhost:3000"
    export PUB_KEY="xxxxx" # an SSH public key to login the VMs
    export PRIV_KEY="xxxxx" # an SSH private key to let blockform login the VMs
    export PASSPHRASE="xxxxx" # the passphrase of the private key

Get, build and run:

    git clone https://github.com/WeTrustPlatform/blockform.git
    cd blockform
    <store the environment variables in the .env file>
    go build && ./blockform

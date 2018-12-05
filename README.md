# BlockForm - Ethereum Blockchain As A Service

BlockForm helps you deploy Ethereum nodes to the cloud and manage them.
It takes care of creating a virtual machine in one of the supported cloud
providers (currently Azure and AWS) and will install geth in the best possible
way.

Blockform is a free software. It is tailored for teams or individuals whose
work depends on a fleet of Ethereum nodes. You have to install your own
instance. So the installation process has been made as simple as possible.

### Features

 - Deploy nodes for public networks
 - Create your own private ethereum network with Proof of Authority
 - Support multiple cloud providers: Azure and AWS
 - RPC endpoints protected by an API key
 - Basic faucet web service for your private network
 - Node data saved to an SSD
 - Basic heathchecks
 - Basic authentication for the UI
 - Works with MetaMask, geth attach, and web3.js
 - Easy to deploy on Heroku

### Usage

Generate your Azure service principal:

    az login
    az ad sp create-for-rbac --sdk-auth

For AWS, go to IAM Management console, and generate an Access Key for your
user in the Security Cretendials tab.

Set environment variables:

    export AZURE_TENANT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_SECRET="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_SUBSCRIPTION_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AWS_ACCESS_KEY_ID="xxxxxxxxxx"
    export AWS_SECRET_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxx"
    export PORT=3000
    export DATABASE_URL="postgres://user:pass@localhost:5432/blockform?sslmode=disable"
    export BASIC_AUTH_USER=xxxxx
    export BASIC_AUTH_PASS=xxxxx
    export PUB_KEY="xxxxx" # your SSH public key to login the VMs

Get, build and run:

    go get github.com/WeTrustPlatform/blockform
    cd $GOPATH/src/github.com/WeTrustPlatform/blockform
    go build && ./blockform

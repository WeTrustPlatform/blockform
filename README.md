# BlockForm - Ethereum Blockchain As A Service

BlockForm helps you deploy ethereum nodes to the cloud and manage them.
It takes care of creating a virtual machine in one of the supported cloud
providers (currently Azure and AWS) and will install geth in the best possible
way.

### Usage

Generate your azure service principal:

    az login
    az ad sp create-for-rbac --sdk-auth

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
    export PUB_KEY="xxxxx" # your public ssh key to login the VMs

Get, build and run:

    go get github.com/WeTrustPlatform/blockform
    cd $GOPATH/src/github.com/WeTrustPlatform/blockform
    go build
    ./blockform

# BlockForm - Ethereum Blockchain As A Service

BlockForm helps you deploy ethereum nodes to the cloud and manage them.
It takes care of creating a virtual machine in one of the supported cloud
providers (currently Azure) and will install geth in the best possible way.

### Usage

Generate your azure service principal:

    az login
    az ad sp create-for-rbac --sdk-auth

Set environment variables:

    export AZURE_TENANT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_CLIENT_SECRET="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export AZURE_SUBSCRIPTION_ID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export PORT=3000
    export DATABASE_URL="postgres://user:pass@localhost:5432/blockform?sslmode=disable"
    export BASIC_AUTH_USER=xxxxx
    export BASIC_AUTH_PASS=xxxxx

Get, build and run:

    go get github.com/WeTrustPlatform/blockform
    cd $GOPATH/src/github.com/WeTrustPlatform/blockform
    go build
    ./blockform

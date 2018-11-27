# BlockForm - Ethereum Blockchain As A Service

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

Build and run:

    go build
    ./blockform

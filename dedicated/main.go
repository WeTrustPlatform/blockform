package dedicated

import (
	"context"
	"log"
	"os"

	"github.com/WeTrustPlatform/blockform/cloudinit"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/WeTrustPlatform/blockform/sshcmd"
)

// Dedicated implements the CloudProvider interface for dedicated servers
type Dedicated struct {
}

// NewDedicated instanciates the package
func NewDedicated() (*Dedicated, error) {
	var dd Dedicated
	return &dd, nil
}

// CreateNode will ssh to the server and run the custom cloud init script
func (dd Dedicated) CreateNode(ctx context.Context, node model.Node, callback func(string, string), onError func(error)) {
	customData := cloudinit.EncodedCustomData(node, "")

	cmd := `sudo apt update && sudo apt install cloud-init && echo '` + customData + `' | base64 -d > config-data && sudo rm -rf /var/lib/cloud/* && sudo cloud-init -f config-data init && sudo cloud-init -f config-data modules && sudo cloud-init -f config-data modules -m final`

	_, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"ubuntu",
		node.DomainName,
		cmd,
	)

	if err != nil {
		log.Println(err)
		onError(err)
		return
	}

	callback(node.DomainName, node.DomainName)
}

// DeleteNode here won't do anything because we are not in a cloud context
func (dd Dedicated) DeleteNode(ctx context.Context, node model.Node, onSuccess func(), onError func(error)) {
	onSuccess()
}

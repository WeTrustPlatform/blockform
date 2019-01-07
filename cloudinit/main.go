package cloudinit

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/WeTrustPlatform/blockform/model"
)

// CustomData returns the raw cloud-init script to initialize a VM.
// The dev argument is the unix device to be used for geth data.
func CustomData(node model.Node, dev string) string {
	var data []byte
	switch node.NetworkID {
	case 1:
		data, _ = ioutil.ReadFile("cloudinit/mainnet.yml")
	case 3:
		data, _ = ioutil.ReadFile("cloudinit/ropsten.yml")
	case 4:
		data, _ = ioutil.ReadFile("cloudinit/rinkeby.yml")
	}

	if node.NetworkType == model.Private {
		data, _ = ioutil.ReadFile("cloudinit/private.yml")
	}

	str := string(data)
	str = strings.Replace(str, "@@API_KEY@@", node.APIKey, -1)
	str = strings.Replace(str, "@@PUB_KEY@@", os.Getenv("PUB_KEY"), -1)
	str = strings.Replace(str, "@@NET_ID@@", fmt.Sprintf("%d", node.NetworkID), -1)
	str = strings.Replace(str, "@@SYNC_MODE@@", fmt.Sprintf("%s", node.SyncMode), -1)
	str = strings.Replace(str, "@@DEVICE@@", dev, -1)

	return str
}

// EncodedCustomData returns the base64 encoded cloud-init script to initialize a VM.
// The dev argument is the unix device to be used for geth data.
func EncodedCustomData(node model.Node, dev string) string {
	data := CustomData(node, dev)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

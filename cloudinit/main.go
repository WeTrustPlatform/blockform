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
	var filename string
	switch node.NetworkID {
	case 1:
		filename = "mainnet.yml"
	case 3:
		filename = "ropsten.yml"
	case 4:
		filename = "rinkeby.yml"
	}

	if node.NetworkType == model.Private {
		filename = "private.yml"
	}

	if node.CloudProvider == "dedicated" {
		filename = "dedicated_" + filename
	}

	data, _ = ioutil.ReadFile("cloudinit/" + filename)

	str := string(data)
	str = strings.Replace(str, "@@NODE_ID@@", fmt.Sprintf("%d", node.ID), -1)
	str = strings.Replace(str, "@@API_KEY@@", node.APIKey, -1)
	str = strings.Replace(str, "@@PUB_KEY@@", os.Getenv("PUB_KEY"), -1)
	str = strings.Replace(str, "@@NET_ID@@", fmt.Sprintf("%d", node.NetworkID), -1)
	str = strings.Replace(str, "@@SYNC_MODE@@", node.SyncMode, -1)
	str = strings.Replace(str, "@@DEVICE@@", dev, -1)
	str = strings.Replace(str, "@@SITE_URL@@", os.Getenv("SITE_URL"), -1)

	return str
}

// EncodedCustomData returns the base64 encoded cloud-init script to initialize a VM.
// The dev argument is the unix device to be used for geth data.
func EncodedCustomData(node model.Node, dev string) string {
	data := CustomData(node, dev)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

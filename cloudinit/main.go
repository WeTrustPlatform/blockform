package cloudinit

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/WeTrustPlatform/blockform/model"
)

// CustomData returns the base64 encoded cloud-init script to initialize a VM
func CustomData(node model.Node) string {
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

	fmt.Println(str)

	return base64.StdEncoding.EncodeToString([]byte(str))
}

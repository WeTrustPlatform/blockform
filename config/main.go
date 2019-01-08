package config

import "github.com/WeTrustPlatform/blockform/model"

// SizeForMode describe the ideal size of hard drives required by geth
// depending on the sync mode chosen
var SizeForMode = map[string]int64{
	model.Full:  2000,
	model.Fast:  200,
	model.Light: 20,
}

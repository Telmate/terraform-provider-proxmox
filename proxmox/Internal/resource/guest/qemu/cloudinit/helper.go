package cloudinit

import "strings"

func trimNameServers(nameServers string) string {
	return strings.ReplaceAll(strings.ReplaceAll(nameServers, " ", ""), ",", "")
}

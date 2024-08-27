package nameservers

import (
	"net/netip"
	"strings"
)

func Split(rawNameServers string) *[]netip.Addr {
	nameServers := make([]netip.Addr, 0)
	if rawNameServers == "" {
		return &nameServers
	}
	nameServerArrays := strings.Split(rawNameServers, " ")
	for _, nameServer := range nameServerArrays {
		nameServerSubArrays := strings.Split(nameServer, ",")
		if len(nameServerSubArrays) > 1 {
			tmpNameServers := make([]netip.Addr, len(nameServerSubArrays))
			for i, e := range nameServerSubArrays {
				tmpNameServers[i], _ = netip.ParseAddr(e)
			}
			nameServers = append(nameServers, tmpNameServers...)
		} else {
			tmpNameServer, _ := netip.ParseAddr(nameServer)
			nameServers = append(nameServers, tmpNameServer)
		}
	}
	return &nameServers
}

func String(nameServers *[]netip.Addr) string {
	if nameServers != nil {
		var rawNameServers string
		for _, nameServer := range *nameServers {
			rawNameServers += " " + nameServer.String()
		}
		if rawNameServers != "" {
			return rawNameServers[1:]
		}
	}
	return ""
}

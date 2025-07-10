package mac

import (
	"net"
)

func Terraform(mac string, id int, currentAdapter map[int]any, key string, params map[string]any) {
	if vv, ok := currentAdapter[id]; ok {
		tfMAC := vv.(map[string]any)[key].(string)
		currentMac, _ := net.ParseMAC(tfMAC)
		if currentMac.String() == mac {
			params[key] = tfMAC
		} else {
			params[key] = mac
		}
	} else {
		params[key] = mac
	}
}

package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the Terraform configuration to the SDK configuration
func SDK(d *schema.ResourceData) (pveAPI.QemuNetworkInterfaces, diag.Diagnostics) {
	networks := make(pveAPI.QemuNetworkInterfaces, AmountNetworkInterfaces)
	for i := 0; i < AmountNetworkInterfaces; i++ {
		networks[pveAPI.QemuNetworkInterfaceID(i)] = pveAPI.QemuNetworkInterface{Delete: true}
	}
	var diags diag.Diagnostics
	for _, e := range d.Get(Root).([]interface{}) {
		networkMap := e.(map[string]interface{})
		id := pveAPI.QemuNetworkInterfaceID(networkMap[schemaID].(int))
		if v, duplicate := networks[id]; duplicate {
			if !v.Delete {
				diags = append(diags, diag.Errorf("Duplicate network interface %s %d", schemaID, id)...)
			}
		}
		tmpMAC, _ := net.ParseMAC(networkMap[schemaMAC].(string))
		mtu := networkMap[schemaMTU].(int)
		var tmpMTU pveAPI.QemuMTU
		model := pveAPI.QemuNetworkModel(networkMap[schemaModel].(string))
		if mtu != 0 {
			if string(pveAPI.QemuNetworkModelVirtIO) == model.String() {
				if mtu == 1 {
					tmpMTU = pveAPI.QemuMTU{Inherit: true}
				} else {
					tmpMTU = pveAPI.QemuMTU{Value: pveAPI.MTU(mtu)}
				}
			} else {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("%s is only supported when model is %s", schemaMTU, pveAPI.QemuNetworkModelVirtIO)})
			}
		}
		rateString, _, _ := strings.Cut(strconv.Itoa(networkMap[schemaRate].(int)), ".")
		rate, _ := strconv.ParseInt(rateString, 10, 64)
		networks[id] = pveAPI.QemuNetworkInterface{
			Bridge:        util.Pointer(networkMap[schemaBridge].(string)),
			Connected:     util.Pointer(!networkMap[schemaLinkDown].(bool)),
			Delete:        false,
			Firewall:      util.Pointer(networkMap[schemaFirewall].(bool)),
			MAC:           &tmpMAC,
			MTU:           &tmpMTU,
			NativeVlan:    util.Pointer(pveAPI.Vlan(networkMap[schemaNativeVlan].(int))),
			Model:         util.Pointer(model),
			MultiQueue:    util.Pointer(pveAPI.QemuNetworkQueue(networkMap[schemaQueues].(int))),
			RateLimitKBps: util.Pointer(pveAPI.GuestNetworkRate(rate * 1000))}
	}
	return networks, diags
}

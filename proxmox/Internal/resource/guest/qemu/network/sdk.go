package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the Terraform configuration to the SDK configuration
func SDK(d *schema.ResourceData) (pxapi.QemuNetworkInterfaces, diag.Diagnostics) {
	networks := make(pxapi.QemuNetworkInterfaces, maximumNetworkInterfaces)
	for i := 0; i < maximumNetworkInterfaces; i++ {
		networks[pxapi.QemuNetworkInterfaceID(i)] = pxapi.QemuNetworkInterface{Delete: true}
	}
	var diags diag.Diagnostics
	for _, e := range d.Get(Root).([]interface{}) {
		networkMap := e.(map[string]interface{})
		id := pxapi.QemuNetworkInterfaceID(networkMap[schemaID].(int))
		if v, duplicate := networks[id]; duplicate {
			if !v.Delete {
				diags = append(diags, diag.Errorf("Duplicate network interface %s %d", schemaID, id)...)
			}
		}
		tmpMAC, _ := net.ParseMAC(networkMap[schemaMAC].(string))
		mtu := networkMap[schemaMTU].(int)
		var tmpMTU pxapi.QemuMTU
		model := pxapi.QemuNetworkModel(networkMap[schemaModel].(string))
		if mtu != 0 {
			if string(pxapi.QemuNetworkModelVirtIO) == model.String() {
				if mtu == 1 {
					tmpMTU = pxapi.QemuMTU{Inherit: false}
				} else {
					tmpMTU = pxapi.QemuMTU{Value: pxapi.MTU(mtu)}
				}
			} else {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("%s is only supported when model is %s", schemaMTU, pxapi.QemuNetworkModelVirtIO)})
			}
		}
		rateString, _, _ := strings.Cut(strconv.Itoa(networkMap[schemaRate].(int)), ".")
		rate, _ := strconv.ParseInt(rateString, 10, 64)
		networks[id] = pxapi.QemuNetworkInterface{
			Bridge:        util.Pointer(networkMap[schemaBridge].(string)),
			Connected:     util.Pointer(!networkMap[schemaLinkDown].(bool)),
			Delete:        false,
			Firewall:      util.Pointer(networkMap[schemaFirewall].(bool)),
			MAC:           &tmpMAC,
			MTU:           &tmpMTU,
			Model:         util.Pointer(model),
			MultiQueue:    util.Pointer(pxapi.QemuNetworkQueue(networkMap[schemaQueues].(int))),
			RateLimitKBps: util.Pointer(pxapi.QemuNetworkRate(rate))}
	}
	return networks, diags
}

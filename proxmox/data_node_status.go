package proxmox

import (
	"context"
	"fmt"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataNodeStatus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataReadNodeStatus,
		Schema: map[string]*schema.Schema{
			"node": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					v, ok := i.(string)
					if !ok {
						return diag.Diagnostics{diag.Diagnostic{
							Severity:      diag.Error,
							Summary:       "Invalid node",
							Detail:        "node must be a string",
							AttributePath: path,
						}}
					}
					if err := pveSDK.NodeName(v).Validate(); err != nil {
						return diag.Diagnostics{diag.Diagnostic{
							Severity:      diag.Error,
							Summary:       "Invalid node",
							Detail:        err.Error(),
							AttributePath: path,
						}}
					}
					return nil
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"uptime": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"mem_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"mem_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataReadNodeStatus(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	nodeName := data.Get("node").(string)

	nodes, err := client.GetNodeList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeData, ok := nodes["data"]
	if !ok {
		return diag.Errorf("unexpected response from Proxmox API: missing 'data' key in node list")
	}

	nodeList, ok := nodeData.([]interface{})
	if !ok {
		return diag.Errorf("unexpected response from Proxmox API: 'data' is not a list")
	}

	var foundNode map[string]interface{}
	for _, n := range nodeList {
		entry, ok := n.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := entry["node"].(string)
		if !ok {
			continue
		}
		if name == nodeName {
			foundNode = entry
			break
		}
	}

	if foundNode == nil {
		return diag.Errorf("node %q not found in Proxmox cluster", nodeName)
	}

	status, _ := foundNode["status"].(string)

	// maintenance is absent for normal nodes; Proxmox sends numeric 1 when set.
	var maintenance bool
	if rawMaint, exists := foundNode["maintenance"]; exists {
		switch v := rawMaint.(type) {
		case bool:
			maintenance = v
		case float64:
			maintenance = v != 0
		}
	}

	// JSON numbers decode as float64; cast to int for whole-number fields.
	uptime, _ := foundNode["uptime"].(float64)
	cpu, _ := foundNode["cpu"].(float64)
	memUsed, _ := foundNode["mem"].(float64)
	memTotal, _ := foundNode["maxmem"].(float64)

	data.SetId(fmt.Sprintf("node/%s", nodeName))
	_ = data.Set("status", status)
	_ = data.Set("maintenance", maintenance)
	_ = data.Set("uptime", int(uptime))
	_ = data.Set("cpu", cpu)
	_ = data.Set("mem_used", int(memUsed))
	_ = data.Set("mem_total", int(memTotal))

	return nil
}

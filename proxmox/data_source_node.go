package proxmox

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNode() *schema.Resource {
	nodeDataSourceDef := &schema.Resource{
		ReadContext: dataSourceNodeRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"level": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maxcpu": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"maxmem": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"mem": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ssl_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uptime": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}

	return nodeDataSourceDef
}

func dataSourceNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return _dataSourceNodeRead(ctx, d, meta)
}

func _dataSourceNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	logger, _ := CreateSubLogger("resource_node_read")

	var diags diag.Diagnostics

	nodeName := d.Get("name").(string)
	logger.Debug().Str("nodeName", nodeName)

	nodeList, err := client.GetNodeList()
	if err != nil {
		d.SetId("")
		return diag.Errorf("unexpected error when trying to get node list: %v", err)
	}

	nodeListData, dataPresent := nodeList["data"]
	if !dataPresent {
		d.SetId("")
		return diag.Errorf("did not get expected format when trying to get node list")
	}

	data := nodeListData.([]interface{})
	for _, nodeDetailsData := range data {
		nodeDetails := nodeDetailsData.(map[string]interface{})

		name := nodeDetails["node"].(string)
		fmt.Printf("CHECKED %s against %s", name, nodeName)
		if name != nodeName {
			continue
		}

		id := nodeDetails["id"].(string)
		d.SetId(id)

		d.Set("status", nodeDetails["status"].(string))
		d.Set("cpu", nodeDetails["cpu"].(float64))
		d.Set("level", nodeDetails["level"].(string))
		d.Set("maxcpu", nodeDetails["maxcpu"].(float64))
		d.Set("maxmem", nodeDetails["maxmem"].(float64))
		d.Set("mem", nodeDetails["mem"].(float64))
		d.Set("ssl_fingerprint", nodeDetails["ssl_fingerprint"].(string))
		d.Set("uptime", nodeDetails["uptime"].(float64))

		return diags
	}

	diags = diag.Errorf("no node found with name \"%s\"", nodeName)
	return diags
}

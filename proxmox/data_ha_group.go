package proxmox

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataHAGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataReadHAGroup,
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restricted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"nofailback": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataReadHAGroup(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {

	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	group, err := client.GetHAGroupByName(ctx, data.Get("group_name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	nodes := group.Nodes
	sort.Strings(nodes)

	data.SetId(group.Group)
	_ = data.Set("nodes", nodes)
	_ = data.Set("type", group.Type)
	_ = data.Set("restricted", group.Restricted)
	_ = data.Set("nofailback", group.NoFailback)
	_ = data.Set("comment", group.Comment)
	return nil
}

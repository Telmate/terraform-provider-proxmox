package proxmox

import (
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataHAGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataReadHAGroup,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataReadHAGroup(d *schema.ResourceData, meta interface{}) (err error) {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	var haGroup *proxmox.HAGroup
	haGroup, err = client.GetHAGroupByName(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.SetId(haGroup.Group)
	_ = d.Set("nodes", haGroup.Nodes)
	_ = d.Set("comment", haGroup.Comment)
	_ = d.Set("type", haGroup.Type)
	return nil
}

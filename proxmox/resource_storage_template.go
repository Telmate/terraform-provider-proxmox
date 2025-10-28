package proxmox

import (
	"context"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStorageTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageTemplateCreate,
		ReadContext:   resourceStorageTemplateRead,
		DeleteContext: resourceStorageTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"pve_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"os_template": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}
}

func _toConfigContent_Template(d *schema.ResourceData) (config pveSDK.ConfigContent_Template, err error) {
	template := d.Get("template").(string)
	storage := d.Get("storage").(string)
	node := d.Get("pve_node").(string)

	config = pveSDK.ConfigContent_Template{
		Node:     node,
		Storage:  storage,
		Template: template,
	}
	err = config.Validate()
	return
}

func resourceStorageTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	config, err := _toConfigContent_Template(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := pveSDK.DownloadLxcTemplate(ctx, client, config); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(config.VolId())
	d.Set("os_template", config.VolId())

	return resourceStorageTemplateRead(ctx, d, meta)
}

func resourceStorageTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	config, err := _toConfigContent_Template(d)
	if err != nil {
		return diag.FromErr(err)
	}

	exists, err := config.Exists(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	if !exists {
		d.SetId("")
	}
	return nil
}

func resourceStorageTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	config, err := _toConfigContent_Template(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := config.Delete(ctx, client); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

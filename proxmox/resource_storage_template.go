package proxmox

import (
	"context"
	"fmt"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	templateContentType = "vztmpl"
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

	template := d.Get("template").(string)
	storage := d.Get("storage").(string)
	node := d.Get("pve_node").(string)

	config := pveSDK.ConfigContent_Template{
		Node:     node,
		Storage:  storage,
		Template: template,
	}
	if err := config.Validate(); err != nil {
		return diag.FromErr(err)
	}

	if err := pveSDK.DownloadLxcTemplate(ctx, client, config); err != nil {
		return diag.FromErr(err)
	}

	volId := fmt.Sprintf("%s:%s/%s", storage, templateContentType, template)
	d.SetId(volId)
	d.Set("os_template", volId)

	return resourceStorageTemplateRead(ctx, d, meta)
}

func resourceStorageTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	var templateFound bool
	storage := d.Get("storage").(string)
	nodeName := pveSDK.NodeName(d.Get("pve_node").(string))

	storageContent, err := client.GetStorageContent(ctx, storage, nodeName)
	if err != nil {
		return diag.FromErr(err)
	}

	contents := storageContent["data"].([]interface{})
	for c := range contents {
		contentMap := contents[c].(map[string]interface{})
		if contentMap["volid"].(string) == d.Id() {
			size := contentMap["size"].(float64)
			d.Set("size", ByteCountIEC(int64(size)))
			templateFound = true
			break
		}
	}

	if !templateFound {
		d.SetId("")
	}

	return nil
}

func resourceStorageTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	storage := strings.SplitN(d.Id(), ":", 2)[0]
	templateURL := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", d.Get("pve_node").(string), storage, d.Id())

	if err := client.Delete(ctx, templateURL); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

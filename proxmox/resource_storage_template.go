package proxmox

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
			"node": {
				Type:     schema.TypeString,
				Description: "Node to install the template on",
				Required: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Description: "Storage to install the template in",
				Required: true,
				ForceNew: true,
			},
			"template": {
				Type:     schema.TypeSet,
				Description: "Information defining the template to download",
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"package": {
							Type: schema.TypeString,
							Description: "Field package in templates pve webui, can be used with the version field",
							Optional: true,
							Default: "",
						},
						"version": {
							Type: schema.TypeString,
							Description: "Field version in templates pve webui",
							Optional: true,
							Default: "",
						},
						"file": {
							Type: schema.TypeString,
							Description: "Exact file name from command `pveam available`",
							Optional: true,
							Default: "",
						},
					},
				},
			},
			"os_template": {
				Type:     schema.TypeString,
				Description: "Template name used to define a container",
				Computed: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}
}

func _parse_template(ctx context.Context, client *pveSDK.Client, d *schema.ResourceData) (template string, err error) {
	node := d.Get("node").(string)
	templateSetList := d.Get("template").(*schema.Set).List()
	if len(templateSetList) == 0 {
		err = errors.New("template is not defined")
		return
	}
	templateSet := templateSetList[0].(map[string]any)

	if file, isSet := templateSet["file"]; isSet && file.(string) != "" {
		// file is one output of `pveam available` like almalinux-9-default_20240911_amd64.tar.xz
		template = file.(string)
	} else if pkg, isSet := templateSet["package"]; isSet && pkg.(string) != "" {
		templatePrefix := pkg.(string)
		if version, isSet := templateSet["version"]; isSet && version.(string) != "" {
			templatePrefix += "_" + version.(string)
		}

		// Will try to find a file that has the prefix {pkg}_{version}
		var availableTemplates *[]pveSDK.TemplateItem
		availableTemplates, err = pveSDK.ListTemplates(ctx, client, node)
		if err != nil {
			return
		}

		for _, availableTemplate := range *availableTemplates {
			if strings.HasPrefix(availableTemplate.Template, templatePrefix) {
				template = availableTemplate.Template
				break
			}
		}

		if template == "" {
			err = fmt.Errorf("Couldn't find a template for package %s", templatePrefix)
		}
	} else {
		err = errors.New("template.file or template.package must be set")
	}
	return
}

func _toConfigContent_Template(ctx context.Context, client *pveSDK.Client, d *schema.ResourceData) (config pveSDK.ConfigContent_Template, err error) {
	storage := d.Get("storage").(string)
	node := d.Get("node").(string)

	template, err := _parse_template(ctx, client, d)
	if err != nil {
		return
	}

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

	config, err := _toConfigContent_Template(ctx, client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := config.Download(ctx, client); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(config.VolId())
	d.Set("os_template", config.VolId())

	return resourceStorageTemplateRead(ctx, d, meta)
}

func resourceStorageTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	config, err := _toConfigContent_Template(ctx, client, d)
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

	config, err := _toConfigContent_Template(ctx, client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := config.Delete(ctx, client); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

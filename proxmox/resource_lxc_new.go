package proxmox

import (
	"context"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/privilege"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/rootmount"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/template"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/name"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/node"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/pool"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var lxcNewResourceDef *schema.Resource

func resourceLxcNew() *schema.Resource {
	lxcNewResourceDef = &schema.Resource{
		CreateContext: resourceLxcNewCreate,
		ReadContext:   resourceLxcNewReadWithLock,
		UpdateContext: resourceLxcNewUpdate,
		DeleteContext: resourceVmQemuDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			name.Root:                  name.Schema(),
			node.RootNode:              node.SchemaNode(schema.Schema{ConflictsWith: []string{node.RootNodes}}, "lxc"),
			node.RootNodes:             node.SchemaNodes("lxc"),
			privilege.RootPrivileged:   privilege.SchemaPrivileged(),
			privilege.RootUnprivileged: privilege.SchemaUnprivileged(),
			rootmount.Root:             rootmount.Schema(),
			template.Root:              template.Schema(),
		},
		Timeouts: resourceTimeouts(),
	}

	return lxcNewResourceDef
}

func resourceLxcNewCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	config, diags := lxcSDK(d)

	// Set the node for the LXC container
	targetNode, err := node.SdkCreate(d)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}
	config.Node = &targetNode

	config.CreateOptions = &pveSDK.LxcCreateOptions{
		OsTemplate: template.SDK(d)}
	config.Privileged = privilege.SDK(d)

	vmr, err := config.Create(ctx, client)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}

	d.SetId(resourceId(targetNode, "lxc", vmr.VmId()))

	return append(diags, resourceLxcNewRead(ctx, d, meta, vmr, client)...)
}

func resourceLxcNewUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pConf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pConf)
	defer lock.unlock()

	client := pConf.Client

	// Get vm reference
	_, _, guestID, err := parseResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return diag.Diagnostics{{
			Summary:  "unexpected error when trying to read and parse the resource: " + err.Error(),
			Severity: diag.Error}}
	}
	var vmr *pveSDK.VmRef
	vmr, err = client.GetVmRefById(ctx, guestID)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error}}
	}

	// create a new config from the resource data
	config, diags := lxcSDK(d)

	// update the targetNode for the LXC container
	var targetNode pveSDK.NodeName
	targetNode, err = node.SdkUpdate(d, vmr.Node())
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}
	config.Node = &targetNode
	config.Pool = util.Pointer(pool.SDK(d))

	if err = config.Update(ctx, vmr, client); err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}

	return append(diags, resourceLxcNewRead(ctx, d, meta, vmr, client)...)
}

func resourceLxcNewReadWithLock(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pConf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pConf)
	defer lock.unlock()

	_, _, guestID, err := parseResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return diag.Diagnostics{{
			Summary:  "unexpected error when trying to read and parse the resource: " + err.Error(),
			Severity: diag.Error}}
	}

	return resourceLxcNewRead(ctx, d, meta, pveSDK.NewVmRef(guestID), pConf.Client)
}

func resourceLxcNewRead(ctx context.Context, d *schema.ResourceData, meta any, vmr *pveSDK.VmRef, client *pveSDK.Client) diag.Diagnostics {
	raw, err := pveSDK.NewConfigLXCFromApi(ctx, vmr, client)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error}}
	}
	config := raw.ALL(*vmr)

	name.Terraform_Unsafe(config.Name, d)
	node.Terraform(*config.Node, d)
	privilege.Terraform(*config.Privileged, d)
	rootmount.Terraform(config.BootMount, d)
	return nil
}

func lxcSDK(d *schema.ResourceData) (pveSDK.ConfigLXC, diag.Diagnostics) {
	var guestName *pveSDK.GuestName
	if v := name.SDK(d); v != "" {
		guestName = &v
	}
	config := pveSDK.ConfigLXC{
		BootMount: rootmount.SDK(d),
		Name:      guestName,
	}
	var diags diag.Diagnostics
	return config, diags
}

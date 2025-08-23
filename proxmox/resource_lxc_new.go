package proxmox

import (
	"context"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/description"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/dns"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/architecture"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/cpu"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/memory"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/mounts"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/networks"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/operatingsystem"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/password"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/privilege"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/rootmount"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/swap"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/template"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/name"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/node"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/pool"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/powerstate"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var lxcNewResourceDef *schema.Resource

func ResourceLxcNew() *schema.Resource {
	lxcNewResourceDef = &schema.Resource{
		CreateContext: resourceLxcNewCreate,
		ReadContext:   resourceLxcNewReadWithLock,
		UpdateContext: resourceLxcNewUpdate,
		DeleteContext: resourceLxcNewDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			architecture.Root:          architecture.Schema(),
			cpu.Root:                   cpu.Schema(),
			description.Root:           description.Schema(),
			dns.Root:                   dns.Schema(),
			memory.Root:                memory.Schema(),
			mounts.RootMount:           mounts.SchemaMount(),
			mounts.RootMounts:          mounts.SchemaMounts(),
			name.Root:                  name.Schema(),
			networks.RootNetwork:       networks.SchemaNetwork(),
			networks.RootNetworks:      networks.SchemaNetworks(),
			node.RootNode:              node.SchemaNode(schema.Schema{ConflictsWith: []string{node.RootNodes}}, "lxc"),
			node.RootNodes:             node.SchemaNodes("lxc"),
			operatingsystem.Root:       operatingsystem.Schema(),
			password.Root:              password.Schema(),
			pool.Root:                  pool.Schema(),
			powerstate.Root:            powerstate.Schema(),
			privilege.RootPrivileged:   privilege.SchemaPrivileged(),
			privilege.RootUnprivileged: privilege.SchemaUnprivileged(),
			rootmount.Root:             rootmount.Schema(),
			swap.Root:                  swap.Schema(),
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

	privileged := privilege.SDK(d)
	config, diags := lxcSDK(privileged, d)
	config.Privileged = &privileged

	// Set the node for the LXC container
	targetNode, err := node.SdkCreate(d)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}
	config.Node = &targetNode

	config.CreateOptions = &pveSDK.LxcCreateOptions{
		OsTemplate:   template.SDK(d),
		UserPassword: password.SDK(d)}

	config.Pool = util.Pointer(pool.SDK(d))

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
	config, diags := lxcSDK(privilege.SDK(d), d)

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

	if err = config.Update(ctx, true, vmr, client); err != nil {
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
	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return diag.Diagnostics{{
			Summary:  err.Error(),
			Severity: diag.Error}}
	}

	raw, err := pveSDK.NewRawConfigLXCFromAPI(ctx, vmr, client)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error}}
	}
	config := raw.Get(*vmr, pveSDK.PowerStateUnknown)

	architecture.Terraform(config.Architecture, d)
	cpu.Terraform(config.CPU, d)
	description.Terraform(config.Description, false, d)
	dns.Terraform(config.DNS, d)
	memory.Terraform(config.Memory, d)
	mounts.Terraform(config.Mounts, d)
	name.Terraform_Unsafe(config.Name, d)
	networks.Terraform(config.Networks, d)
	node.Terraform(*config.Node, d)
	operatingsystem.Terraform(config.OperatingSystem, d)
	pool.Terraform(config.Pool, d)
	powerstate.Terraform(guestStatus.GetState(), d)
	privilege.Terraform(*config.Privileged, d)
	rootmount.Terraform(config.BootMount, d)
	swap.Terraform(config.Swap, d)
	return nil
}

func resourceLxcNewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return guestDelete(ctx, d, meta, "LXC")
}

func lxcSDK(privilidged bool, d *schema.ResourceData) (pveSDK.ConfigLXC, diag.Diagnostics) {
	var guestName *pveSDK.GuestName
	if v := name.SDK(d); v != "" {
		guestName = &v
	}
	config := pveSDK.ConfigLXC{
		BootMount:   rootmount.SDK(privilidged, d),
		CPU:         cpu.SDK(d),
		DNS:         dns.SDK(d),
		Description: description.SDK(false, d),
		Memory:      memory.SDK(d),
		Name:        guestName,
		State:       powerstate.SDK(d),
		Swap:        swap.SDK(d),
	}
	var diags, tmpDiags diag.Diagnostics
	config.Networks, diags = networks.SDK(d)
	if diags.HasError() {
		return config, diags
	}
	config.Mounts, tmpDiags = mounts.SDK(privilidged, d)
	diags = append(diags, tmpDiags...)
	return config, diags
}

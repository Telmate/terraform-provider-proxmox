package proxmox

import (
	"context"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/clone"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/description"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/dns"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/guestid"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/architecture"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/cpu"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/features"
	tags "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/lxc_tags"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/memory"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/mounts"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/networks"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/operatingsystem"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/password"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/privilege"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/rootmount"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/ssh_public_keys"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/swap"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/template"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/name"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/node"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/pool"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/powerstate"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/reboot"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/startatnodeboot"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/startupshutdown"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/id"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var lxcNewResourceDef *schema.Resource

func resourceLxcGuest() *schema.Resource {
	lxcNewResourceDef = &schema.Resource{
		CreateContext: resourceLxcGuestCreate,
		ReadContext:   resourceLxcGuestReadWithLock,
		UpdateContext: resourceLxcGuestUpdate,
		DeleteContext: resourceLxcGuestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: reboot.CustomizeDiff(),

		Schema: map[string]*schema.Schema{
			architecture.Root:            architecture.Schema(),
			clone.Root:                   clone.Schema(),
			cpu.Root:                     cpu.Schema(),
			description.Root:             description.Schema(),
			dns.Root:                     dns.Schema(),
			features.Root:                features.Schema(),
			guestid.Root:                 guestid.Schema(),
			memory.Root:                  memory.Schema(),
			mounts.RootMount:             mounts.SchemaMount(),
			mounts.RootMounts:            mounts.SchemaMounts(),
			name.Root:                    name.Schema(),
			networks.RootNetwork:         networks.SchemaNetwork(),
			networks.RootNetworks:        networks.SchemaNetworks(),
			node.RootNode:                node.SchemaNode(schema.Schema{ConflictsWith: []string{node.RootNodes}}, "lxc"),
			node.RootNodes:               node.SchemaNodes("lxc"),
			operatingsystem.Root:         operatingsystem.Schema(),
			password.Root:                password.Schema(),
			pool.Root:                    pool.Schema(),
			powerstate.Root:              powerstate.Schema(),
			privilege.RootPrivileged:     privilege.SchemaPrivileged(),
			privilege.RootUnprivileged:   privilege.SchemaUnprivileged(),
			reboot.RootAutomatic:         reboot.SchemaAutomatic(),
			reboot.RootAutomaticSeverity: reboot.SchemaAutomaticSeverity(),
			reboot.RootRequired:          reboot.SchemaRequired(),
			rootmount.Root:               rootmount.Schema(),
			ssh_public_keys.Root:         ssh_public_keys.Schema(),
			startatnodeboot.Root:         startatnodeboot.Schema(),
			startupshutdown.Root:         startupshutdown.Schema(),
			swap.Root:                    swap.Schema(),
			tags.Root:                    tags.Schema(),
			template.Root:                template.Schema(),
		},
		Timeouts: resourceTimeouts(),
	}

	return lxcNewResourceDef
}

func resourceLxcGuestCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	diags := lxcGuestWarning()

	client := pconf.Client

	privileged := privilege.SDK(d)
	config, tmpDiags := lxcSDK(privileged, d)
	diags = append(diags, tmpDiags...)
	if diags.HasError() {
		return diags
	}
	config.ID = guestid.SDK(d)
	config.Privileged = &privileged

	// Set the node for the LXC container
	targetNode, err := node.SdkCreate(d)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}

	var vmr *pveSDK.VmRef

	cloneGuest := clone.SDK(d, clone.Settings{
		ID:   config.ID,
		Name: config.Name,
		Node: targetNode,
		Pool: config.Pool})
	if cloneGuest != nil {
		var cloneRef *pveSDK.VmRef
		cloneRef, err = guestGetSourceVmr(ctx, client, cloneGuest.Name, cloneGuest.ID, targetNode, pveSDK.GuestLxc, clone.Root+" { "+clone.SchemaName+" }", clone.Root+" { "+clone.SchemaID+" }")
		if err != nil {
			return append(diags, diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error})
		}
		vmr, err = cloneRef.CloneLxc(ctx, cloneGuest.Target, client)
		if err != nil {
			return append(diags, diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error})
		}
		d.SetId(id.Guest{
			ID:   vmr.VmId(),
			Node: targetNode,
			Type: id.GuestLxc}.String())
		err = config.Update(ctx, true, vmr, client)
		if err != nil {
			return append(diags, diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error})
		}
	} else {
		config.Node = &targetNode
		config.CreateOptions = &pveSDK.LxcCreateOptions{
			OsTemplate:    template.SDK(d),
			PublicSSHkeys: ssh_public_keys.SDK(d),
			UserPassword:  password.SDK(d)}
		config.Pool = util.Pointer(pool.SDK(d))
		vmr, err = config.Create(ctx, client)
		if err != nil {
			return append(diags, diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error})
		}
		d.SetId(id.Guest{
			ID:   vmr.VmId(),
			Node: targetNode,
			Type: id.GuestLxc}.String())
	}

	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client)...)
}

func resourceLxcGuestUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pConf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pConf)
	defer lock.unlock()

	client := pConf.Client

	diags := lxcGuestWarning()

	// Get vm reference
	var resourceID id.Guest
	err := resourceID.Parse(d.Id())
	if err != nil {
		d.SetId("")
		return append(diags, diag.Diagnostic{
			Summary:  "unexpected error when trying to read and parse the resource: " + err.Error(),
			Severity: diag.Error})
	}
	var vmr *pveSDK.VmRef
	vmr, err = client.GetVmRefById(ctx, resourceID.ID)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}

	// create a new config from the resource data
	config, tmpDiags := lxcSDK(privilege.SDK(d), d)
	diags = append(diags, tmpDiags...)
	if diags.HasError() {
		return diags
	}

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

	if err = config.Update(ctx, reboot.GetAutomatic(d), vmr, client); err != nil {
		if err.Error() == "<this should be the reboot error>" { // TODO catch the error but we need upstream support for that
			return append(diags, reboot.ErrorLxc(d))
		}
		return append(diags, diag.Diagnostic{
			Summary:  err.Error(),
			Severity: diag.Error})
	}

	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client)...)
}

func resourceLxcGuestReadWithLock(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pConf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pConf)
	defer lock.unlock()

	diags := lxcGuestWarning()

	var resourceID id.Guest
	if err := resourceID.Parse(d.Id()); err != nil {
		d.SetId("")
		return append(diags, diag.Diagnostic{
			Summary:  "unexpected error when trying to read and parse the resource: " + err.Error(),
			Severity: diag.Error})
	}

	client := pConf.Client

	ok, err := resourceID.ID.Exists(ctx, client)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	if !ok {
		return append(diags, resourceDriftDeletionDiagnostic(d))
	}

	vmr := pveSDK.NewVmRef(resourceID.ID)
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client)...)
}

func resourceLxcGuestRead(ctx context.Context, d *schema.ResourceData, vmr *pveSDK.VmRef, client *pveSDK.Client) diag.Diagnostics {
	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return diag.Diagnostics{{
			Summary:  err.Error(),
			Severity: diag.Error}}
	}

	var raw pveSDK.RawConfigLXC
	var pending bool
	raw, pending, err = pveSDK.NewActiveRawConfigLXCFromApi(ctx, vmr, client)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Summary:  err.Error(),
				Severity: diag.Error}}
	}
	reboot.SetRequired(pending, d)

	d.SetId(id.Guest{
		ID:   vmr.VmId(),
		Node: vmr.Node(),
		Type: id.GuestLxc}.String())

	config := raw.Get(*vmr, pveSDK.PowerStateUnknown)

	architecture.Terraform(config.Architecture, d)
	cpu.Terraform(config.CPU, d)
	description.Terraform(config.Description, false, d)
	dns.Terraform(config.DNS, d)
	features.Terraform(config.Features, d)
	guestid.Terraform(config.ID, d)
	memory.Terraform(config.Memory, d)
	mounts.Terraform(config.Mounts, d)
	name.Terraform_Unsafe(config.Name, d)
	if err = networks.Terraform(config.Networks, d); err != nil {
		return diag.FromErr(err)
	}
	node.Terraform(*config.Node, d)
	operatingsystem.Terraform(config.OperatingSystem, d)
	pool.Terraform(config.Pool, d)
	powerstate.Terraform(guestStatus.GetState(), d)
	privilege.Terraform(*config.Privileged, d)
	rootmount.Terraform(config.BootMount, d)
	startatnodeboot.Terraform(*config.StartAtNodeBoot, d)
	startupshutdown.Terraform(config.StartupShutdown, d)
	swap.Terraform(config.Swap, d)
	tags.Terraform(config.Tags, d)
	return nil
}

func resourceLxcGuestDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return guestDelete(ctx, d, meta, "LXC")
}

func lxcSDK(privilidged bool, d *schema.ResourceData) (pveSDK.ConfigLXC, diag.Diagnostics) {
	var guestName *pveSDK.GuestName
	if v := name.SDK(d); v != "" {
		guestName = &v
	}
	config := pveSDK.ConfigLXC{
		BootMount:       rootmount.SDK(privilidged, d),
		CPU:             cpu.SDK(d),
		DNS:             dns.SDK(d),
		Description:     description.SDK(false, d),
		Features:        features.SDK(privilidged, d),
		Memory:          memory.SDK(d),
		Name:            guestName,
		StartAtNodeBoot: util.Pointer(startatnodeboot.SDK(d)),
		StartupShutdown: startupshutdown.SDK(d),
		State:           powerstate.SDK(d),
		Swap:            swap.SDK(d),
		Tags:            tags.SDK(d),
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

func lxcGuestWarning() diag.Diagnostics {
	return diag.Diagnostics{{
		Detail:   "The LXC Guest resource is experimental. The schema and functionality may change in future releases without a major version bump.",
		Summary:  "LXC Guest resource is experimental",
		Severity: diag.Warning}}
}

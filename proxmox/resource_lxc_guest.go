package proxmox

import (
	"context"
	"time"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/clone"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/description"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/dns"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/guestid"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/ip"
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
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/wait"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/id"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var lxcNewResourceDef *schema.Resource

const (
	schemaNetworkTimeout = "network_timeout"
)

func resourceLxcGuest() *schema.Resource {
	lxcNewResourceDef = &schema.Resource{
		CreateContext: resourceLxcGuestCreate,
		ReadContext:   resourceLxcGuestReadWithLock,
		UpdateContext: resourceLxcGuestUpdate,
		DeleteContext: resourceLxcGuestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			networks.CustomizeDiff(),
			reboot.CustomizeDiff(),
		),

		Schema: map[string]*schema.Schema{
			architecture.Root:            architecture.Schema(),
			clone.Root:                   clone.Schema(),
			cpu.Root:                     cpu.Schema(),
			description.Root:             description.Schema(),
			dns.Root:                     dns.Schema(),
			features.Root:                features.Schema(),
			guestid.Root:                 guestid.Schema(),
			ip.RootLxcV4:                 ip.SchemaV4(),
			ip.RootLxcV6:                 ip.SchemaV6(),
			ip.RootSkipV4:                ip.SchemaSkipV4(),
			ip.RootSkipV6:                ip.SchemaSkipV6(),
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
			powerstate.Root:              powerstate.Schema(schema.Schema{Default: powerstate.Default}),
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
			wait.Root:                    wait.Schema(),
			schemaNetworkTimeout: {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Timeout in seconds to keep trying to obtain an IP address.",
				ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
					if i.(int) > 0 {
						return nil
					}
					return diag.Errorf(schemaNetworkTimeout + " must be greater than 0")
				}}},
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
	clientNew := pconf.NewClient

	version, err := client.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	privileged := privilege.SDK(d)
	config, tmpDiags := lxcSDK(privileged, version, d)
	diags = append(diags, tmpDiags...)
	if diags.HasError() {
		return diags
	}
	config.ID = guestid.SDK(d)
	config.Pool = new(pool.SDK(d))
	config.Privileged = &privileged

	// Set the node for the LXC container
	var targetNode pveSDK.NodeName
	targetNode, err = node.SdkCreate(d)
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
		cloneRef, err = guestGetSourceVmr(ctx, clientNew.Guest, cloneGuest.Name, cloneGuest.ID, targetNode, pveSDK.GuestLxc, clone.Root+" { "+clone.SchemaName+" }", clone.Root+" { "+clone.SchemaID+" }")
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

	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client, &version, true)...)
}

func resourceLxcGuestUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pConf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pConf)
	defer lock.unlock()

	client := pConf.Client

	diags := lxcGuestWarning()

	version, err := client.Version(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get vm reference
	var resourceID id.Guest
	err = resourceID.Parse(d.Id())
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
	config, tmpDiags := lxcSDK(privilege.SDK(d), version, d)
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

	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client, &version, true)...)
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
	return append(diags, resourceLxcGuestRead(ctx, d, vmr, client, nil, false)...)
}

func resourceLxcGuestRead(ctx context.Context, d *schema.ResourceData, vmr *pveSDK.VmRef, client *pveSDK.Client, version *pveSDK.Version, waitForAgent bool) diag.Diagnostics {
	newClient := client.New()
	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return diag.FromErr(err)
	}

	var raw pveSDK.RawConfigLXC
	var pending bool
	raw, pending, err = pveSDK.NewActiveRawConfigLXCFromApi(ctx, vmr, client)
	if err != nil {
		return diag.FromErr(err)
	}
	reboot.SetRequired(pending, d)

	d.SetId(id.Guest{
		ID:   vmr.VmId(),
		Node: vmr.Node(),
		Type: id.GuestLxc}.String())

	var poolPtr *pveSDK.PoolName
	if v := vmr.Pool(); v != "" {
		poolPtr = &v
	}

	config := raw.Get(poolPtr, pveSDK.PowerStateUnknown)

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
	node.Terraform(vmr.Node(), d)
	operatingsystem.Terraform(config.OperatingSystem, d)
	pool.Terraform(config.Pool, d)
	state := guestStatus.GetState()
	powerstate.Terraform(state, false, d)
	privilege.Terraform(*config.Privileged, d)
	rootmount.Terraform(config.BootMount, d)
	startatnodeboot.Terraform(*config.StartAtNodeBoot, d)
	startupshutdown.Terraform(config.StartupShutdown, d)
	swap.Terraform(config.Swap, d)
	tags.Terraform(config.Tags, d)

	var diags diag.Diagnostics
	if state == pveSDK.PowerStateRunning {
		if len(config.Networks) == 0 {
			return nil
		}
		conn := &connectionInfo{
			SkipIPv4: d.Get(ip.RootSkipV4).(bool),
			SkipIPv6: d.Get(ip.RootSkipV6).(bool),
		}
		var timeout time.Duration
		if waitForAgent {
			timeout = time.Duration(d.Get(schemaNetworkTimeout).(int)) * time.Second
		}
		tmpDiags := lxcGetIP(ctx, client, newClient, conn, config.Networks, *vmr, timeout, wait.GetDuration(d), version)
		if len(tmpDiags) > 0 {
			diags = append(diags, tmpDiags...)
		}
		d.Set(ip.RootLxcV4, conn.IPs.IPv4)
		d.Set(ip.RootLxcV6, conn.IPs.IPv6)
	}
	return diags
}

func resourceLxcGuestDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return guestDelete(ctx, d, meta, "LXC")
}

func lxcSDK(privilidged bool, version pveSDK.Version, d *schema.ResourceData) (pveSDK.ConfigLXC, diag.Diagnostics) {
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
		StartAtNodeBoot: new(startatnodeboot.SDK(d)),
		StartupShutdown: startupshutdown.SDK(d),
		State:           powerstate.SDK(powerstate.LegacyFalse, d),
		Swap:            swap.SDK(d),
		Tags:            tags.SDK(d),
	}
	var diags, tmpDiags diag.Diagnostics
	config.Networks, diags = networks.SDK(version.Encode(), d)
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

func lxcGetIP(ctx context.Context, client *pveSDK.Client, newClient pveSDK.ClientNew, conn *connectionInfo, network pveSDK.LxcNetworks, vmr pveSDK.VmRef, retryDuration, retryInterval time.Duration, version *pveSDK.Version) diag.Diagnostics {
	var name pveSDK.LxcNetworkName
	for i := range networks.NetworksAmount {
		if v, ok := network[pveSDK.LxcNetworkID(i)]; ok {
			if v.IPv4 != nil && v.IPv4.Address != nil {
				conn.SkipIPv4 = true
				conn.IPs.IPv4 = v.IPv4.Address.String()
			}
			if v.IPv6 != nil && v.IPv6.Address != nil {
				conn.SkipIPv6 = true
				conn.IPs.IPv6 = v.IPv6.Address.String()
			}
			name = *v.Name
			break
		}
	}
	if !conn.SkipIPv4 || !conn.SkipIPv6 {
		if version == nil {
			tmpVersion, err := client.Version(ctx)
			if err != nil {
				return diag.FromErr(err)
			}
			version = &tmpVersion
		}
		if version.Encode() < (pveSDK.Version{Major: 9, Minor: 1}.Encode()) {
			return nil
		}
		endTime := time.Now().Add(retryDuration)
		for {
			info, err := newClient.LxcGuest.ReadNetworkInterfaceInfo(ctx, vmr)
			if err != nil {
				return diag.FromErr(err)
			}
			if interfaceInfo, ok := info.SelectName(name); ok {
				conn.parsePrimaryIPs(interfaceInfo.GetIpAddresses())
				if conn.hasRequiredIP() {
					return nil
				}
			}
			if !time.Now().Before(endTime) {
				return conn.agentDiagnostics(schemaNetworkTimeout, "", "")
			}
			time.Sleep(retryInterval)
		}
	}
	return nil
}

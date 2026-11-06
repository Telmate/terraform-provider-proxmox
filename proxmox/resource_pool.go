package proxmox

import (
	"context"
	"fmt"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const schemaPoolComment = "comment"

var poolResourceDef *schema.Resource

func resourcePool() *schema.Resource {
	poolResourceDef = &schema.Resource{
		CreateContext: resourcePoolCreate,
		ReadContext:   resourcePoolRead,
		UpdateContext: resourcePoolUpdate,
		DeleteContext: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"poolid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			schemaPoolComment: {
				Type:     schema.TypeString,
				Default:  defaultDescription,
				Optional: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}

	return poolResourceDef
}

func resourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.NewClient
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	poolid := d.Get("poolid").(string)

	err := client.Pool.Create(ctx, pveSDK.ConfigPool{
		Name:    pveSDK.PoolName(poolid),
		Comment: util.Pointer(d.Get("comment").(string)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(clusterResourceId("pools", poolid))

	return diag.FromErr(_resourcePoolRead(ctx, d, meta))
}

func resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return diag.FromErr(_resourcePoolRead(ctx, d, meta))
}

func _resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.NewClient

	_, poolID, err := parseClusterResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unexpected error when trying to read and parse resource id: %v", err)
	}

	logger, _ := CreateSubLogger("resource_pool_read")
	logger.Info().Str("poolid", poolID).Msg("Reading configuration for poolid")

	rawPool, err := client.Pool.Read(ctx, pveSDK.PoolName(poolID))
	if err != nil {
		exists, errExists := client.Pool.Exists(ctx, pveSDK.PoolName(poolID))
		if errExists == nil && !exists {
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(clusterResourceId("pools", poolID))
	d.Set("poolid", poolID)
	d.Set("comment", rawPool.GetComment())

	// DEBUG print the read result
	logger.Debug().Str("poolid", poolID).Msgf("Finished pool read resulting in data: '%+v'", rawPool.Get())
	return nil
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	logger, _ := CreateSubLogger("resource_pool_update")

	client := pconf.NewClient
	_, poolID, err := parseClusterResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	logger.Info().Str("poolid", poolID).Msg("Starting update of the Pool resource")

	if d.HasChange("comment") {
		err := client.Pool.Update(ctx, pveSDK.ConfigPool{
			Name:    pveSDK.PoolName(poolID),
			Comment: util.Pointer(d.Get("comment").(string)),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(_resourcePoolRead(ctx, d, meta))
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.NewClient
	_, poolID, err := parseClusterResourceId(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = client.Pool.Delete(ctx, pveSDK.PoolName(poolID)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

package proxmox

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var poolResourceDef *schema.Resource

func resourcePool() *schema.Resource {
	poolResourceDef = &schema.Resource{
		Create: resourcePoolCreate,
		Read:   resourcePoolRead,
		Update: resourcePoolUpdate,
		Delete: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"poolid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}

	return poolResourceDef
}

func resourcePoolCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	poolid := d.Get("poolid").(string)
	comment := d.Get("comment").(string)

	err := client.CreatePool(poolid, comment)
	if err != nil {
		return err
	}

	d.SetId(clusterResourceId("pools", poolid))

	return _resourcePoolRead(d, meta)
}

func resourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return _resourcePoolRead(d, meta)
}

func _resourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	_, poolID, err := parseClusterResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unexpected error when trying to read and parse resource id: %v", err)
	}

	logger, _ := CreateSubLogger("resource_pool_read")
	logger.Info().Str("poolid", poolID).Msg("Reading configuration for poolid")

	poolInfo, err := client.GetPoolInfo(poolID)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.SetId(clusterResourceId("pools", poolID))
	d.Set("comment", "")
	if poolInfo["comment"] != nil {
		d.Set("comment", poolInfo["comment"].(string))
	}

	// DEBUG print the read result
	logger.Debug().Str("poolid", poolID).Msgf("Finished pool read resulting in data: '%+v'", poolInfo)
	return nil
}

func resourcePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	logger, _ := CreateSubLogger("resource_pool_update")

	client := pconf.Client
	_, poolID, err := parseClusterResourceId(d.Id())
	if err != nil {
		return err
	}

	logger.Info().Str("poolid", poolID).Msg("Starting update of the Pool resource")

	if d.HasChange("comment") {
		nextComment := d.Get("comment").(string)
		err := client.UpdatePoolComment(poolID, nextComment)
		if err != nil {
			return err
		}
	}

	return _resourcePoolRead(d, meta)
}

func resourcePoolDelete(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	_, poolID, err := parseClusterResourceId(d.Id())

	if err != nil {
		return err
	}
	err = client.DeletePool(poolID)
	if err != nil {
		return err
	}

	return nil
}

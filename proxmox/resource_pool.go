package proxmox

import (
	"fmt"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var poolResourceDef *schema.Resource

func resourcePool() *schema.Resource {
	*pxapi.Debug = true

	poolResourceDef = &schema.Resource{
		Create: resourcePoolCreate,
		Read:   resourcePoolRead,
		Update: resourcePoolUpdate,
		Delete: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
		return fmt.Errorf("Unexpected error when trying to read and parse resource id: %v", err)
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
	if poolInfo["data"].(map[string]interface{})["comment"] != nil {
		d.Set("comment", poolInfo["data"].(map[string]interface{})["comment"].(string))
	}

	// DEBUG print the read result
	logger.Debug().Str("poolid", poolID).Msgf("Finished pool read resulting in data: '%+v'", poolInfo["data"])
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

	err = client.DeletePool(poolID)
	if err != nil {
		return err
	}

	return nil
}

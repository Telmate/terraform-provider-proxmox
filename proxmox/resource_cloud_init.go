package proxmox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/kdomanski/iso9660"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	isoContentType = "iso"
)

func resourceCloudInitDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudInitDiskCreate,
		ReadContext:   resourceCloudInitDiskRead,
		DeleteContext: resourceCloudInitDiskDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"meta_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vendor_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_config": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sha256": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func createCloudInitISO(metaData, userData, vendorData, networkConfig string) (io.Reader, string, error) {
	isoWriter, err := iso9660.NewWriter()
	if err != nil {
		return nil, "", err
	}
	defer isoWriter.Cleanup()

	if metaData != "" {
		err = isoWriter.AddFile(strings.NewReader(metaData), "meta-data")
		if err != nil {
			return nil, "", err
		}
	}

	if userData != "" {
		err = isoWriter.AddFile(strings.NewReader(userData), "user-data")
		if err != nil {
			return nil, "", err
		}
	}

	if vendorData != "" {
		err = isoWriter.AddFile(strings.NewReader(vendorData), "vendor-data")
		if err != nil {
			return nil, "", err
		}
	}

	if networkConfig != "" {
		err = isoWriter.AddFile(strings.NewReader(networkConfig), "network-config")
		if err != nil {
			return nil, "", err
		}
	}

	var b bytes.Buffer
	err = isoWriter.WriteTo(&b, "cidata")
	if err != nil {
		return nil, "", err
	}

	// Calculate the ISO sha256 sum
	sum := fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))

	return bytes.NewReader(b.Bytes()), sum, nil
}

func resourceCloudInitDiskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pconf := m.(*providerConfiguration)
	client := pconf.Client

	r, sum, err := createCloudInitISO(d.Get("meta_data").(string), d.Get("user_data").(string), d.Get("vendor_data").(string), d.Get("network_config").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	fileName := fmt.Sprintf("tf-ci-%s.iso", d.Get("name").(string))
	err = client.Upload(d.Get("pve_node").(string), d.Get("storage").(string), isoContentType, fileName, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("sha256", sum)

	// The volume ID is the storage name and the file name
	volId := fmt.Sprintf("%s:%s/%s", d.Get("storage").(string), isoContentType, fileName)
	d.SetId(volId)

	return resourceCloudInitDiskRead(ctx, d, m)
}

func resourceCloudInitDiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pconf := m.(*providerConfiguration)
	client := pconf.Client

	var isoFound bool
	pveNode := d.Get("pve_node").(string)
	vmRef := &proxmox.VmRef{}
	vmRef.SetNode(pveNode)
	vmRef.SetVmType("qemu")
	storageContent, err := client.GetStorageContent(vmRef, d.Get("storage").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	contents := storageContent["data"].([]interface{})
	for c := range contents {
		storageContentMap := contents[c].(map[string]interface{})
		if storageContentMap["volid"].(string) == d.Id() {
			size := storageContentMap["size"].(float64)
			d.Set("size", ByteCountIEC(int64(size)))
			isoFound = true
			break
		}
	}
	if !isoFound {
		// ISO not found so we (re)create it
		d.SetId("")
	}

	return nil
}

func resourceCloudInitDiskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pconf := m.(*providerConfiguration)
	client := pconf.Client

	storage := strings.SplitN(d.Id(), ":", 2)[0]
	isoURL := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", d.Get("pve_node").(string), storage, d.Id())
	err := client.Delete(isoURL)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

package proxmox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStorageIso() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageIsoCreate,
		ReadContext:   resourceStorageIsoRead,
		DeleteContext: resourceStorageIsoDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"checksum_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pve_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}
}

func resourceStorageIsoCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)

	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	url := d.Get("url").(string)
	fileName := d.Get("filename").(string)
	storage := d.Get("storage").(string)
	node := d.Get("pve_node").(string)

	client := pconf.Client
	file, err := os.CreateTemp(os.TempDir(), fileName)
	if err != nil {
		return diag.FromErr(err)
	}
	err = _downloadFile(url, file)
	if err != nil {
		return diag.FromErr(err)
	}
	file.Seek(0, 0)
	defer file.Close()
	err = client.Upload(ctx, node, storage, isoContentType, fileName, file)
	if err != nil {
		return diag.FromErr(err)
	}
	volId := fmt.Sprintf("%s:%s/%s", storage, isoContentType, fileName)
	d.SetId(volId)

	return resourceStorageIsoRead(ctx, d, meta)
}

func _downloadFile(url string, file *os.File) error {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func resourceStorageIsoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	var isoFound bool
	storageContent, err := client.GetStorageContent(ctx, d.Get("storage").(string), pveSDK.NodeName(d.Get("pve_node").(string)))
	if err != nil {
		return diag.FromErr(err)
	}
	contents := storageContent["data"].([]interface{})
	for c := range contents {
		contentMap := contents[c].(map[string]interface{})
		if contentMap["volid"].(string) == d.Id() {
			size := contentMap["size"].(float64)
			d.Set("size", ByteCountIEC(int64(size)))
			isoFound = true
			break
		}
	}

	if !isoFound {
		d.SetId("")
	}

	return nil
}

func resourceStorageIsoDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	storage := strings.SplitN(d.Id(), ":", 2)[0]
	isoURL := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", d.Get("pve_node").(string), storage, d.Id())
	err := client.Delete(ctx, isoURL)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

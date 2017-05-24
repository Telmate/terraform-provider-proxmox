package proxmox

import (
	"fmt"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
	"strconv"
	"sync"
)

type providerConfiguration struct {
	Client *pxapi.Client
}

func Provider() *schema.Provider {
	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"pm_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_USER", nil),
				Description: "username, maywith with @pam",
			},
			"pm_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_PASS", nil),
				Description: "secret",
				Sensitive:   true,
			},
			"pm_api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_URL", nil),
				Description: "https://host.fqdn:8006/api2/json",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"proxmox_vm_qemu": resourceVmQemu(),
			// TODO - storage_iso
			// TODO - bridge
			// TODO - vm_qemu_template
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client, err := getClient(d.Get("pm_api_url").(string), d.Get("pm_user").(string), d.Get("pm_password").(string))
	if err != nil {
		return nil, err
	}
	return &providerConfiguration{
		Client: client,
	}, nil
}

func getClient(pm_api_url string, pm_user string, pm_password string) (*pxapi.Client, error) {
	client, _ := pxapi.NewClient(pm_api_url, nil, nil)
	err := client.Login(pm_user, pm_password)
	if err != nil {
		return nil, err
	}
	return client, nil
}

var mutex = &sync.Mutex{}
var maxVmId = 0

func nextVmId(client *pxapi.Client) (nextId int, err error) {
	mutex.Lock()
	if maxVmId == 0 {
		maxVmId, err = pxapi.MaxVmId(client)
		if err != nil {
			return 0, err
		}
	}
	maxVmId++
	nextId = maxVmId
	mutex.Unlock()
	return nextId, nil
}

func resourceId(targetNode string, resType string, vmId int) string {
	return fmt.Sprintf("%s/%s/%d", targetNode, resType, vmId)
}

var rxRsId = regexp.MustCompile("([^/]+)/([^/]+)/(\\d+)")

func parseResourceId(resId string) (targetNode string, resType string, vmId int, err error) {
	idMatch := rxRsId.FindStringSubmatch(resId)
	if idMatch == nil {
		err = fmt.Errorf("Invalid resource id: %s", resId)
	}
	targetNode = idMatch[1]
	resType = idMatch[2]
	vmId, err = strconv.Atoi(idMatch[3])
	return
}

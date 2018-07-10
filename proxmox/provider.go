package proxmox

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

type providerConfiguration struct {
	Client          *pxapi.Client
	MaxParallel     int
	CurrentParallel int
	MaxVmId         int
	Mutex           *sync.Mutex
	Cond            *sync.Cond
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
			"pm_parallel": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
			"pm_tls_insecure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
	client, err := getClient(d.Get("pm_api_url").(string), d.Get("pm_user").(string), d.Get("pm_password").(string), d.Get("pm_tls_insecure").(bool))
	if err != nil {
		return nil, err
	}
	var mut sync.Mutex
	return &providerConfiguration{
		Client:          client,
		MaxParallel:     d.Get("pm_parallel").(int),
		CurrentParallel: 0,
		MaxVmId:         0,
		Mutex:           &mut,
		Cond:            sync.NewCond(&mut),
	}, nil
}

func getClient(pm_api_url string, pm_user string, pm_password string, pm_tls_insecure bool) (*pxapi.Client, error) {
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !pm_tls_insecure {
		tlsconf = nil
	}
	client, _ := pxapi.NewClient(pm_api_url, nil, tlsconf)
	err := client.Login(pm_user, pm_password)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func nextVmId(pconf *providerConfiguration) (nextId int, err error) {
	pconf.Mutex.Lock()
	if pconf.MaxVmId == 0 {
		pconf.MaxVmId, err = pxapi.MaxVmId(pconf.Client)
		if err != nil {
			return 0, err
		}
	}
	pconf.MaxVmId++
	nextId = pconf.MaxVmId
	pconf.Mutex.Unlock()
	return nextId, nil
}

func pmParallelBegin(pconf *providerConfiguration) {
	pconf.Mutex.Lock()
	for pconf.CurrentParallel >= pconf.MaxParallel {
		pconf.Cond.Wait()
	}
	pconf.CurrentParallel++
	pconf.Mutex.Unlock()
}

func pmParallelEnd(pconf *providerConfiguration) {
	pconf.Mutex.Lock()
	pconf.CurrentParallel--
	pconf.Cond.Signal()
	pconf.Mutex.Unlock()
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

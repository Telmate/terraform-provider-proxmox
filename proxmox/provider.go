package proxmox

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type providerConfiguration struct {
	Client                             *pxapi.Client
	MaxParallel                        int
	CurrentParallel                    int
	MaxVMID                            int
	Mutex                              *sync.Mutex
	Cond                               *sync.Cond
	LogFile                            string
	LogLevels                          map[string]string
	DangerouslyIgnoreUnknownAttributes bool
}

// Provider - Terrafrom properties for proxmox
func Provider() *schema.Provider {
	pmOTPprompt := schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		DefaultFunc: schema.EnvDefaultFunc("PM_OTP", ""),
		Description: "OTP 2FA code (if required)",
	}
	if os.Getenv("PM_OTP_PROMPT") == "1" {
		pmOTPprompt = schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("PM_OTP", nil),
			Description: "OTP 2FA code (if required)",
		}
	}
	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"pm_user": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_USER", nil),
				Description: "Username e.g. myuser or myuser@pam",
			},
			"pm_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_PASS", nil),
				Description: "Password to authenticate into proxmox",
				Sensitive:   true,
			},
			"pm_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_URL", ""),
				ValidateFunc: func(v interface{}, k string) (warns []string, errs []error) {
					value := v.(string)

					if value == "" {
						errs = append(errs, fmt.Errorf("you must specify an endpoint for the Proxmox Virtual Environment API (valid: https://host:port)"))
						return
					}

					_, err := url.ParseRequestURI(value)

					if err != nil {
						errs = append(errs, fmt.Errorf("you must specify an endpoint for the Proxmox Virtual Environment API (valid: https://host:port)"))
						return
					}

					return
				},
				Description: "https://host.fqdn:8006/api2/json",
			},
			"pm_api_token_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_TOKEN_ID", nil),
				Description: "API TokenID e.g. root@pam!mytesttoken",
			},
			"pm_api_token_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_TOKEN_SECRET", nil),
				Description: "The secret uuid corresponding to a TokenID",
				Sensitive:   true,
			},
			"pm_parallel": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
			"pm_tls_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_TLS_INSECURE", true), //we assume it's a lab!
				Description: "By default, every TLS connection is verified to be secure. This option allows terraform to proceed and operate on servers considered insecure. For example if you're connecting to a remote host and you do not have the CA cert that issued the proxmox api url's certificate.",
			},
			"pm_http_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_HTTP_HEADERS", nil),
				Description: "Set custom http headers e.g. Key,Value,Key1,Value1",
			},
			"pm_log_enable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable provider logging to get proxmox API logs",
			},
			"pm_log_levels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configure the logging level to display; trace, debug, info, warn, etc",
			},
			"pm_log_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "terraform-plugin-proxmox.log",
				Description: "Write logs to this specific file",
			},
			"pm_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_TIMEOUT", 1200),
				Description: "How many seconds to wait for operations for both provider and api-client, default is 20m",
			},
			"pm_dangerously_ignore_unknown_attributes": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_DANGEROUSLY_IGNORE_UNKNOWN_ATTRIBUTES", false),
				Description: "By default this provider will exit if an unknown attribute is found. This is to prevent the accidential destruction of VMs or Data when something in the proxmox API has changed/updated and is not confirmed to work with this provider. Set this to true at your own risk. It may allow you to proceed in cases when the provider refuses to work, but be aware of the danger in doing so.",
			},
			"pm_debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_DEBUG", false),
				Description: "Enable or disable the verbose debug output from proxmox api",
			},
			"pm_proxy_server": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_PROXY", nil),
				Description: "Proxy Server passed to Api client(useful for debugging). Syntax: http://proxy:port",
			},
			"pm_otp": &pmOTPprompt,
		},

		ResourcesMap: map[string]*schema.Resource{
			"proxmox_vm_qemu":         resourceVmQemu(),
			"proxmox_lxc":             resourceLxc(),
			"proxmox_lxc_disk":        resourceLxcDisk(),
			"proxmox_pool":            resourcePool(),
			"proxmox_cloud_init_disk": resourceCloudInitDisk(),
			"proxmox_storage_iso":     resourceStorageIso(),
			// TODO - proxmox_bridge
			// TODO - proxmox_vm_qemu_template
		},

		DataSourcesMap: map[string]*schema.Resource{
			"proxmox_ha_groups": DataHAGroup(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client, err := getClient(
		d.Get("pm_api_url").(string),
		d.Get("pm_user").(string),
		d.Get("pm_password").(string),
		d.Get("pm_api_token_id").(string),
		d.Get("pm_api_token_secret").(string),
		d.Get("pm_otp").(string),
		d.Get("pm_tls_insecure").(bool),
		d.Get("pm_http_headers").(string),
		d.Get("pm_timeout").(int),
		d.Get("pm_debug").(bool),
		d.Get("pm_proxy_server").(string),
	)
	if err != nil {
		return nil, err
	}

	//permission check
	minimumPermissions := []string{
		"Datastore.AllocateSpace",
		"Datastore.Audit",
		"Pool.Allocate",
		"Sys.Audit",
		"Sys.Console",
		"Sys.Modify",
		"VM.Allocate",
		"VM.Audit",
		"VM.Clone",
		"VM.Config.CDROM",
		"VM.Config.Cloudinit",
		"VM.Config.CPU",
		"VM.Config.Disk",
		"VM.Config.HWType",
		"VM.Config.Memory",
		"VM.Config.Network",
		"VM.Config.Options",
		"VM.Migrate",
		"VM.Monitor",
		"VM.PowerMgmt",
	}
	var id string
	if result, getok := d.GetOk("pm_api_token_id"); getok {
		id = result.(string)
		id = strings.Split(id, "!")[0]
	} else if result, getok := d.GetOk("pm_user"); getok {
		id = result.(string)
	}
	userID, err := pxapi.NewUserID(id)
	if err != nil {
		return nil, err
	}
	permlist, err := client.GetUserPermissions(userID, "/")
	if err != nil {
		return nil, err
	}
	sort.Strings(permlist)
	sort.Strings(minimumPermissions)
	permDiff := permissions_check(permlist, minimumPermissions)
	if len(permDiff) == 0 {
		// look to see what logging we should be outputting according to the provider configuration
		logLevels := make(map[string]string)
		for logger, level := range d.Get("pm_log_levels").(map[string]interface{}) {
			levelAsString, ok := level.(string)
			if ok {
				logLevels[logger] = levelAsString
			} else {
				return nil, fmt.Errorf("invalid logging level %v for %v. Be sure to use a string", level, logger)
			}
		}

		// actually configure logging
		// note that if enable is false here, the configuration will squash all output
		ConfigureLogger(
			d.Get("pm_log_enable").(bool),
			d.Get("pm_log_file").(string),
			logLevels,
		)

		var mut sync.Mutex
		return &providerConfiguration{
			Client:                             client,
			MaxParallel:                        d.Get("pm_parallel").(int),
			CurrentParallel:                    0,
			MaxVMID:                            -1,
			Mutex:                              &mut,
			Cond:                               sync.NewCond(&mut),
			LogFile:                            d.Get("pm_log_file").(string),
			LogLevels:                          logLevels,
			DangerouslyIgnoreUnknownAttributes: d.Get("pm_dangerously_ignore_unknown_attributes").(bool),
		}, nil
	}
	err = fmt.Errorf("permissions for user/token %s are not sufficient, please provide also the following permissions that are missing: %v", userID.ToString(), permDiff)
	return nil, err
}

func getClient(pm_api_url string,
	pm_user string,
	pm_password string,
	pm_api_token_id string,
	pm_api_token_secret string,
	pm_otp string,
	pm_tls_insecure bool,
	pm_http_headers string,
	pm_timeout int,
	pm_debug bool,
	pm_proxy_server string) (*pxapi.Client, error) {

	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !pm_tls_insecure {
		tlsconf = nil
	}

	var err error

	if pm_password != "" && pm_api_token_secret != "" {
		err = fmt.Errorf("password and API token secret both exist, choose one or the other")
	}

	if pm_password == "" && pm_api_token_secret == "" {
		err = fmt.Errorf("password and API token do not exist, one of these must exist")
	}

	if strings.Contains(pm_user, "!") && pm_password != "" {
		err = fmt.Errorf("you appear to be using an API TokenID username with your password")
	}

	if !strings.Contains(pm_api_token_id, "!") {
		err = fmt.Errorf("your API TokenID username should contain a !, check your API credentials")
	}

	client, _ := pxapi.NewClient(pm_api_url, nil, pm_http_headers, tlsconf, pm_proxy_server, pm_timeout)
	*pxapi.Debug = pm_debug

	// User+Pass authentication
	if pm_user != "" && pm_password != "" {
		err = client.Login(pm_user, pm_password, pm_otp)
	}

	// API authentication
	if pm_api_token_id != "" && pm_api_token_secret != "" {
		// Unsure how to get an err for this
		client.SetAPIToken(pm_api_token_id, pm_api_token_secret)
	}

	if err != nil {
		return nil, err
	}
	return client, nil
}

func nextVmId(pconf *providerConfiguration) (nextId int, err error) {
	pconf.Mutex.Lock()
	defer pconf.Mutex.Unlock()
	pconf.MaxVMID, err = pconf.Client.GetNextID(pconf.MaxVMID + 1)
	if err != nil {
		return 0, err
	}
	nextId = pconf.MaxVMID
	return nextId, nil
}

type pmApiLockHolder struct {
	locked bool
	pconf  *providerConfiguration
}

func (lock *pmApiLockHolder) lock() {
	if lock.locked {
		return
	}
	lock.locked = true
	pconf := lock.pconf
	pconf.Mutex.Lock()
	for pconf.CurrentParallel >= pconf.MaxParallel {
		pconf.Cond.Wait()
	}
	pconf.CurrentParallel++
	pconf.Mutex.Unlock()
}

func (lock *pmApiLockHolder) unlock() {
	if !lock.locked {
		return
	}
	lock.locked = false
	pconf := lock.pconf
	pconf.Mutex.Lock()
	pconf.CurrentParallel--
	pconf.Cond.Signal()
	pconf.Mutex.Unlock()
}

func pmParallelBegin(pconf *providerConfiguration) *pmApiLockHolder {
	lock := &pmApiLockHolder{
		pconf:  pconf,
		locked: false,
	}
	lock.lock()
	return lock
}

func resourceId(targetNode string, resType string, vmId int) string {
	return fmt.Sprintf("%s/%s/%d", targetNode, resType, vmId)
}

func parseResourceId(resId string) (targetNode string, resType string, vmId int, err error) {
	// create a logger for this function
	logger, _ := CreateSubLogger("parseResourceId")

	if !rxRsId.MatchString(resId) {
		return "", "", -1, fmt.Errorf("invalid resource format: %s. Must be <node>/<type>/<vmid>", resId)
	}
	idMatch := rxRsId.FindStringSubmatch(resId)
	targetNode = idMatch[1]
	resType = idMatch[2]
	vmId, err = strconv.Atoi(idMatch[3])
	if err != nil {
		logger.Info().Str("error", err.Error()).Msgf("failed to get vmId")
	}
	return
}

func clusterResourceId(resType string, resId string) string {
	return fmt.Sprintf("%s/%s", resType, resId)
}

func parseClusterResourceId(resId string) (resType string, id string, err error) {
	if !rxClusterRsId.MatchString(resId) {
		return "", "", fmt.Errorf("invalid resource format: %s. Must be <type>/<resourceid>", resId)
	}
	idMatch := rxClusterRsId.FindStringSubmatch(resId)
	return idMatch[1], idMatch[2], nil
}

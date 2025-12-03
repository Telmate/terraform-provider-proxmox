package proxmox

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/validator"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	schemaPmUser                               = "pm_user"
	schemaPmPassword                           = "pm_password"
	schemaPmApiUrl                             = "pm_api_url"
	schemaPmApiTokenID                         = "pm_api_token_id"
	schemaPmApiTokenSecret                     = "pm_api_token_secret"
	schemaPmParallel                           = "pm_parallel"
	schemaPmTlsInsecure                        = "pm_tls_insecure"
	schemaPmHttpHeaders                        = "pm_http_headers"
	schemaPmLogEnable                          = "pm_log_enable"
	schemaPmLogLevels                          = "pm_log_levels"
	schemaPmLogFile                            = "pm_log_file"
	schemaPmTimeout                            = "pm_timeout"
	schemaPmDangerouslyIgnoreUnknownAttributes = "pm_dangerously_ignore_unknown_attributes"
	schemaPmDebug                              = "pm_debug"
	schemaPmProxyServer                        = "pm_proxy_server"
	schemaMinimumPermissionCheck               = "pm_minimum_permission_check"
	schemaMinimumPermissionList                = "pm_minimum_permission_list"
	schemaPmOTP                                = "pm_otp"
)

type providerConfiguration struct {
	Client                             *pveSDK.Client
	MaxParallel                        int
	CurrentParallel                    int
	MaxGuestID                         pveSDK.GuestID
	Mutex                              *sync.Mutex
	Cond                               *sync.Cond
	LogFile                            string
	LogLevels                          map[string]string
	DangerouslyIgnoreUnknownAttributes bool
}

type sessionCache struct {
	Ticket string `json:"ticket"`
	Csrf   string `json:"csrf"`
}

func ticketCachePath(cacheKey string) string {
	sum := sha256.Sum256([]byte(cacheKey))
	return filepath.Join(os.TempDir(), fmt.Sprintf("proxmox-ticket-%x", sum[:8]))
}

func loadCachedTicket(path string) (ticket string, csrf string, err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var sc sessionCache
	if err := json.Unmarshal(bytes, &sc); err != nil {
		return "", "", err
	}
	return sc.Ticket, sc.Csrf, nil
}

func saveCachedTicket(path string, ticket string, csrf string) error {
	data, err := json.Marshal(sessionCache{Ticket: ticket, Csrf: csrf})
	if err != nil {
		return err
	}
	// ensure directory exists (os.TempDir should exist already)
	return os.WriteFile(path, data, 0o600)
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
			schemaPmUser: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_USER", nil),
				Description: "Username e.g. myuser or myuser@pam",
			},
			schemaPmPassword: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_PASS", nil),
				Description: "Password to authenticate into proxmox",
				Sensitive:   true,
			},
			schemaPmApiUrl: {
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
			schemaPmApiTokenID: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_TOKEN_ID", nil),
				Description: "API TokenID e.g. root@pam!mytesttoken",
			},
			schemaPmApiTokenSecret: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_API_TOKEN_SECRET", nil),
				Description: "The secret uuid corresponding to a TokenID",
				Sensitive:   true,
			},
			schemaPmParallel: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
					v, ok := i.(int)
					if !ok {
						return diag.Errorf(validator.ErrorUint, k)
					}
					if v < 1 {
						return diag.Errorf(schemaPmParallel + " must be greater than 0")
					}
					return nil
				},
			},
			schemaPmTlsInsecure: {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_TLS_INSECURE", false), // we assume it's a production environment.
				Description: "By default, every TLS connection is verified to be secure. This option allows terraform to proceed and operate on servers considered insecure. For example if you're connecting to a remote host and you do not have the CA cert that issued the proxmox api url's certificate.",
			},
			schemaPmHttpHeaders: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_HTTP_HEADERS", nil),
				Description: "Set custom http headers e.g. Key,Value,Key1,Value1",
			},
			schemaPmLogEnable: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable provider logging to get proxmox API logs",
			},
			schemaPmLogLevels: {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configure the logging level to display; trace, debug, info, warn, etc",
			},
			schemaPmLogFile: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "terraform-plugin-proxmox.log",
				Description: "Write logs to this specific file",
			},
			schemaPmTimeout: {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_TIMEOUT", 1200),
				Description: "How many seconds to wait for operations for both provider and api-client, default is 20m",
			},
			schemaPmDangerouslyIgnoreUnknownAttributes: {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_DANGEROUSLY_IGNORE_UNKNOWN_ATTRIBUTES", false),
				Description: "By default this provider will exit if an unknown attribute is found. This is to prevent the accidential destruction of VMs or Data when something in the proxmox API has changed/updated and is not confirmed to work with this provider. Set this to true at your own risk. It may allow you to proceed in cases when the provider refuses to work, but be aware of the danger in doing so.",
			},
			schemaPmDebug: {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_DEBUG", false),
				Description: "Enable or disable the verbose debug output from proxmox api",
			},
			schemaPmProxyServer: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PM_PROXY", nil),
				Description: "Proxy Server passed to Api client(useful for debugging). Syntax: http://proxy:port",
			},
			schemaMinimumPermissionCheck: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true},
			schemaMinimumPermissionList: {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString}},
			schemaPmOTP: &pmOTPprompt,
		},

		ResourcesMap: map[string]*schema.Resource{
			"proxmox_vm_qemu":         resourceVmQemu(),
			"proxmox_lxc":             resourceLxc(),
			"proxmox_lxc_disk":        resourceLxcDisk(),
			"proxmox_lxc_guest":       resourceLxcGuest(),
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

func providerConfigure(d *schema.ResourceData) (any, error) {
	client, err := getClient(
		d.Get(schemaPmApiUrl).(string),
		d.Get(schemaPmUser).(string),
		d.Get(schemaPmPassword).(string),
		d.Get(schemaPmApiTokenID).(string),
		d.Get(schemaPmApiTokenSecret).(string),
		d.Get(schemaPmOTP).(string),
		d.Get(schemaPmTlsInsecure).(bool),
		d.Get(schemaPmHttpHeaders).(string),
		d.Get(schemaPmTimeout).(int),
		d.Get(schemaPmDebug).(bool),
		d.Get(schemaPmProxyServer).(string),
	)
	if err != nil {
		return nil, err
	}

	minimumPermissions := []string{
		"Datastore.AllocateSpace",
		"Datastore.Audit",
		"Pool.Allocate",
		"Pool.Audit",
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
		"VM.PowerMgmt",
	}
	if v, ok := d.GetOk(schemaMinimumPermissionList); ok {
		rawArray := v.([]any)
		minimumPermissions = make([]string, len(rawArray))
		for i := range rawArray {
			minimumPermissions[i] = rawArray[i].(string)
		}
	}

	if d.Get(schemaMinimumPermissionCheck).(bool) { // permission check
		var id string
		if result, ok := d.GetOk(schemaPmApiTokenID); ok {
			id = result.(string)
			id = strings.Split(id, "!")[0]
		} else if result, ok := d.GetOk(schemaPmUser); ok {
			id = result.(string)
		}
		userID, err := pveSDK.NewUserID(id)
		if err != nil {
			return nil, err
		}
		permList, err := client.GetUserPermissions(context.Background(), userID, "/")
		if err != nil {
			return nil, err
		}
		sort.Strings(permList)
		sort.Strings(minimumPermissions)
		permDiff := permissions_check(permList, minimumPermissions)
		if len(permDiff) != 0 {
			return nil, fmt.Errorf("permissions for user/token "+userID.ToString()+" are not sufficient, please provide also the following permissions that are missing: %v", permDiff)
		}
	}

	// look to see what logging we should be outputting according to the provider configuration
	logLevels := make(map[string]string)
	for logger, level := range d.Get(schemaPmLogLevels).(map[string]any) {
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
		d.Get(schemaPmLogEnable).(bool),
		d.Get(schemaPmLogFile).(string),
		logLevels)

	var mut sync.Mutex
	return &providerConfiguration{
		Client:                             client,
		MaxParallel:                        d.Get(schemaPmParallel).(int),
		CurrentParallel:                    0,
		MaxGuestID:                         0,
		Mutex:                              &mut,
		Cond:                               sync.NewCond(&mut),
		LogFile:                            d.Get(schemaPmLogFile).(string),
		LogLevels:                          logLevels,
		DangerouslyIgnoreUnknownAttributes: d.Get(schemaPmDangerouslyIgnoreUnknownAttributes).(bool),
	}, nil
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
	pm_proxy_server string) (*pveSDK.Client, error) {

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

	client, _ := pveSDK.NewClient(pm_api_url, nil, pm_http_headers, tlsconf, pm_proxy_server, pm_timeout)
	*pveSDK.Debug = pm_debug

	// Try to reuse a cached ticket (plan/apply are separate processes) only when using OTP
	cachePath, loadedFromCache := "", false
	if pm_user != "" && pm_password != "" && pm_otp != "" {
		cacheKey := fmt.Sprintf("%s|%s|%s|%s", pm_api_url, pm_user, pm_password, pm_otp)
		cachePath = ticketCachePath(cacheKey)
		if ticket, csrf, loadErr := loadCachedTicket(cachePath); loadErr == nil && ticket != "" {
			client.SetTicket(ticket, csrf)
			client.Username = pm_user
			// Validate the ticket is still usable; otherwise fall back to login
			if _, pingErr := client.GetVersion(context.Background()); pingErr == nil {
				loadedFromCache = true
				return client, nil
			}
		}
	}

	// User+Pass authentication
	if pm_user != "" && pm_password != "" {
		err = client.Login(context.Background(), pm_user, pm_password, pm_otp)
	}

	// API authentication
	if pm_api_token_id != "" && pm_api_token_secret != "" {
		// Unsure how to get an err for this
		client.SetAPIToken(pm_api_token_id, pm_api_token_secret)
	}

	if err != nil {
		return nil, err
	}

	// Persist the ticket so the next provider process (apply after plan) can reuse it
	// without consuming another OTP
	if cachePath != "" && !loadedFromCache {
		if ticket, csrf, ok := client.SessionTokens(); ok {
			_ = saveCachedTicket(cachePath, ticket, csrf)
		}
	}

	return client, nil
}

func nextVmId(pconf *providerConfiguration) (nextId pveSDK.GuestID, err error) {
	pconf.Mutex.Lock()
	defer pconf.Mutex.Unlock()
	nextId, err = pconf.Client.GetNextID(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	pconf.MaxGuestID = nextId

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

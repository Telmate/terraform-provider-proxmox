package main

import (
	"flag"

	"github.com/Telmate/terraform-provider-proxmox/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {

	var debugMode bool
	var pluginPath string

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.StringVar(&pluginPath, "registry", "registry.terraform.io/telmate/proxmox", "specify path, useful for local debugging")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: func() *schema.Provider {
		return proxmox.Provider()
	}, Debug: debugMode, ProviderAddr: pluginPath}

	plugin.Serve(opts)
}

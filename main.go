package main

import (
	"context"
	"flag"
	"github.com/claranet/terraform-provider-zabbix/zabbix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"log"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	p := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return zabbix.Provider()
		},
	}
	if debug {
		err := plugin.Debug(context.Background(), "citizen.devops.atypon.com/gcp/zabbix", p)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(p)

}

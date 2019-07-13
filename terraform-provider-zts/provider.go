package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type ProviderResources struct {
	ControlServer string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"control_server": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The address to the control server",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zts_secrets":  resourceSecrets(),
			"zts_gpg_keys": resourceGpgKeys(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	return &ProviderResources{
		ControlServer: d.Get("control_server").(string),
	}, nil
}

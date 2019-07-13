package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"etcd_host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The etcd server to store config",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zts_secrets":  resourceSecrets(),
			"zts_gpg_keys": resourceGpgKeys(),
		},
	}
}

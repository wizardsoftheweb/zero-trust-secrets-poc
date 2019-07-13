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
				DefaultFunc: schema.EnvDefaultFunc("ZTS_ETCD_HOST", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zts_secrets": resourceServer(),
		},
	}
}

package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSecrets() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecretsCreate,
		Read:   resourceSecretsRead,
		Update: resourceSecretsUpdate,
		Delete: resourceSecretsDelete,

		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
func resourceSecretsCreate(d *schema.ResourceData, m interface{}) error {
	return resourceSecretsRead(d, m)
}

func resourceSecretsRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceSecretsUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceSecretsRead(d, m)
}

func resourceSecretsDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

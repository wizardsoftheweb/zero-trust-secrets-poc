package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGpgKeys() *schema.Resource {
	return &schema.Resource{
		Create: resourceGpgKeysCreate,
		Read:   resourceGpgKeysRead,
		Update: resourceGpgKeysUpdate,
		Delete: resourceGpgKeysDelete,

		Schema: map[string]*schema.Schema{
			"directory": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
func resourceGpgKeysCreate(d *schema.ResourceData, m interface{}) error {
	return resourceGpgKeysRead(d, m)
}

func resourceGpgKeysRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceGpgKeysUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceGpgKeysRead(d, m)
}

func resourceGpgKeysDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

package main

import (
	"path/filepath"

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
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Directory to store the generated keys",
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					return validateFileObject(true, i, s)
				},
				StateFunc: resolvePath,
			},
			"batch": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name-Real",
						},
						"email": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name-Email",
						},
						"comment": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name-Comment",
						},
					},
				},
			},
			"pub_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".pubring.gpg",
				Description: "The desired pub key basename",
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".secring.gpg",
				Description: "The desired secret key basename",
			},
		},
	}
}

func resolvePath(objectPathSchemaString interface{}) string {
	objectPath, _ := objectPathSchemaString.(string)
	return filepath.Clean(objectPath)
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

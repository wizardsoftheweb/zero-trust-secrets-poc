package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
)

type GpgKeyType int

const (
	GpgKeyTypePub GpgKeyType = iota
	GpgKeyTypeSecret
)

func resourceSecrets() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecretsCreate,
		Read:   resourceSecretsRead,
		Update: resourceSecretsUpdate,
		Delete: resourceSecretsDelete,

		Schema: map[string]*schema.Schema{
			"etcd_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path in etcd to store the file",
			},
			"pub_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    false,
				Description: "The path to the pub key to use",
				DefaultFunc: func() (interface{}, error) {
					return getDefaultGpgKeyFileName(GpgKeyTypePub), nil
				},
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					return validateFileObject(false, i, s)
				},
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    false,
				Description: "The path to the secret key to use",
				DefaultFunc: func() (interface{}, error) {
					return getDefaultGpgKeyFileName(GpgKeyTypeSecret), nil
				},
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					return validateFileObject(false, i, s)
				},
			},
		},
	}
}

func getGpgKeyFileName(directory string, keyType GpgKeyType) string {
	if GpgKeyTypeSecret == keyType {
		return fmt.Sprintf("%s/.secring.gpg", directory)
	}
	return fmt.Sprintf("%s/.pubring.gpg", directory)
}

func getDefaultGpgKeyFileName(keyType GpgKeyType) string {
	cwd, _ := os.Getwd()
	return getGpgKeyFileName(cwd, keyType)
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

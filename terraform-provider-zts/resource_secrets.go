package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strings"

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
			"etcd_host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The etcd server to store config",
			},
			"etcd_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path in etcd to store the file",
			},
			"secret_count": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "The number of random strings to generate",
			},
			"pub_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
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
				Optional:    true,
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
	providerResources := m.(*ProviderResources)
	request := &RandoRequest{
		KvHosts: []string{
			d.Get("etcd_host").(string),
		},
		KvKey:  d.Get("etcd_key").(string),
		Count:  d.Get("secret_count").(int),
		PubKey: loadPubKey(d.Get("pub_key").(string)),
	}
	contents := GenerateSecrets(
		providerResources.ControlServer,
		request,
	)

	d.SetId(
		fmt.Sprintf(
			"%x",
			sha256.Sum256(
				[]byte(strings.Join(contents, ",")),
			),
		),
	)
	return nil
}

func resourceSecretsRead(d *schema.ResourceData, m interface{}) error {
	configManager := *newConfigManager(
		[]string{
			d.Get("etcd_host").(string),
		},
		d.Get("pub_key").(string),
	)
	contents, err := configManager.Get(d.Get("etcd_key").(string))
	log.Println(string(contents))
	return err
}

func resourceSecretsUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceSecretsRead(d, m)
}

func resourceSecretsDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

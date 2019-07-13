package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	goetcd "github.com/coreos/etcd/client"
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
				ForceNew:    true,
				Description: "The etcd server to store config",
			},
			"etcd_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
			"random_secrets": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The generated secrets",
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
	_ = d.Set("random_secrets", contents)
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
	var etcdHost, etcdKey, secretKey string
	if d.HasChange("etcd_host") {
		old, _ := d.GetChange("etcd_host")
		etcdHost = old.(string)
	} else {
		etcdHost = d.Get("etcd_host").(string)
	}
	if d.HasChange("etcd_key") {
		old, _ := d.GetChange("etcd_key")
		etcdKey = old.(string)
	} else {
		etcdKey = d.Get("etcd_key").(string)
	}
	if d.HasChange("secret_key") {
		old, _ := d.GetChange("secret_key")
		secretKey = old.(string)
	} else {
		secretKey = d.Get("secret_key").(string)
	}
	configManager := *newConfigManager(
		[]string{
			etcdHost,
		},
		secretKey,
	)
	contents, err := configManager.Get(etcdKey)
	if nil != err {
		return err
	}
	var parsedContents RandoResponse
	err = json.Unmarshal(contents, &parsedContents)
	if nil != err {
		return err
	}
	log.Println(string(contents))
	err = d.Set("random_secrets", parsedContents.Secrets)
	return err
}

func resourceSecretsUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceSecretsCreate(d, m)
}

func resourceSecretsDelete(d *schema.ResourceData, m interface{}) error {
	client, err := goetcd.New(goetcd.Config{
		Endpoints: []string{
			d.Get("etcd_host").(string),
		},
	})
	if nil != err {
		return err
	}
	keysApi := goetcd.NewKeysAPI(client)
	_, err = keysApi.Delete(
		context.TODO(),
		d.Get("etcd_key").(string),
		&goetcd.DeleteOptions{},
	)
	return err
}

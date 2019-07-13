package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGpgKeys() *schema.Resource {
	return &schema.Resource{
		Create: resourceGpgKeysCreate,
		Read:   resourceGpgKeysRead,
		Delete: resourceGpgKeysDelete,

		Schema: map[string]*schema.Schema{
			"directory": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "Directory to store the generated keys",
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					return validateFileObject(true, i, s)
				},
				StateFunc: resolvePath,
			},
			"batch": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Name-Real",
						},
						"email": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Name-Email",
						},
						"comment": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Name-Comment",
						},
					},
				},
			},
			"pub_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     ".pubring.gpg",
				Description: "The desired pub key basename",
			},
			"computed_pub_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full pub key file path",
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     ".secring.gpg",
				Description: "The desired secret key basename",
			},
			"computed_secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full secret key file path",
			},
			"gpg_key_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The GPG key ID",
			},
		},
	}
}

const batchFormatString = `
%%echo Generating a configuration OpenPGP key
%%no-protection
Key-Type: default
Subkey-Type: default
Name-Real: %s
Name-Comment: %s
Name-Email: %s
Expire-Date: 0
%%commit
%%echo done
`

var keyIdPattern, _ = regexp.Compile(`^\s+[^\s]*?\s*$`)

func resolvePath(objectPathSchemaString interface{}) string {
	objectPath, _ := objectPathSchemaString.(string)
	absPath, _ := filepath.Abs(objectPath)
	return absPath
}

func getValueWithChange(d *schema.ResourceData, key string, useOld bool) interface{} {
	if d.HasChange(key) {
		oldValue, newValue := d.GetChange(key)
		if useOld {
			return oldValue
		}
		return newValue
	}
	return d.Get(key)
}

func initializeResourceCall(d *schema.ResourceData, useOld bool) error {
	directory := getValueWithChange(d, "directory", useOld).(string)
	pubKeyBasename := getValueWithChange(d, "pub_key", useOld).(string)
	secretKeyBasename := getValueWithChange(d, "secret_key", useOld).(string)
	properDirectory := resolvePath(directory)
	err := d.Set("directory", properDirectory)
	if nil != err {
		return err
	}
	err = d.Set(
		"computed_pub_key",
		filepath.Clean(filepath.Join(properDirectory, pubKeyBasename)),
	)
	if nil != err {
		return err
	}
	err = d.Set(
		"computed_secret_key",
		filepath.Clean(filepath.Join(properDirectory, secretKeyBasename)),
	)
	if nil != err {
		return err
	}
	return nil
}

func buildBatchFileContents(name, comment, email string) string {
	return fmt.Sprintf(batchFormatString, name, comment, email)
}

func buildBatchFile(directory, contents string) (string, error) {
	fileName := filepath.Join(directory, "gpg-batch")
	err := ioutil.WriteFile(
		fileName,
		[]byte(contents),
		0644,
	)
	return fileName, err
}

func getBatchComponents(d *schema.ResourceData) (string, string, string, error) {
	var name, comment, email string
	var batchMap map[string]interface{}
	var ok bool
	batch := d.Get("batch").(*schema.Set).List()
	if 1 == len(batch) {
		batchMap, ok = batch[0].(map[string]interface{})
		if !ok {
			return name, comment, email, fmt.Errorf("unable to parse %v", batch)
		}
	} else {
		return name, comment, email, fmt.Errorf("multiple batches found")
	}
	for key, value := range batchMap {
		parsedKey := key
		switch parsedKey {
		case "name":
			name = value.(string)
		case "email":
			email = value.(string)
		case "comment":
			comment = value.(string)
		default:
			return name, comment, email, fmt.Errorf("batch did not contain an error")
		}
	}
	return name, comment, email, nil
}

func execBatchGeneration(d *schema.ResourceData) error {
	var fileName string
	directory := d.Get("directory").(string)
	name, comment, email, err := getBatchComponents(d)
	if nil != err {
		return err
	}
	contents := buildBatchFileContents(name, comment, email)
	fileName, err = buildBatchFile(directory, contents)
	if nil != err {
		return fmt.Errorf("unable to build file: %s", err.Error())
	}
	defer func() {
		_ = os.Remove(fileName)
	}()
	oldWd, err := os.Getwd()
	if nil != err {
		return fmt.Errorf("unable to get current working directory: %s", err.Error())
	}
	err = os.Chdir(directory)
	if nil != err {
		return fmt.Errorf("unable to change directory: %s", err.Error())
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()
	commands := []string{
		"gpg2",
		"--batch",
		"--armor",
		"--gen-key",
		fileName,
	}
	response := execCmd(commands...)
	if !response.Succeeded() {
		return fmt.Errorf("unable to execute batch: %s", response.String())
	}
	return nil
}

func constructGpgKeySearchString(name, comment, email string) string {
	return fmt.Sprintf("%s (%s) <%s>", name, comment, email)
}

func determineKeyId(d *schema.ResourceData, fromKeyring bool) error {
	var keyId, searchString string
	name, comment, email, err := getBatchComponents(d)
	if nil != err {
		return err
	}
	command := []string{
		"gpg2",
	}
	if fromKeyring {
		searchString = constructGpgKeySearchString(name, comment, email)
		command = append(command, "--list-keys")
		command = append(command, searchString)
	} else {
		command = append(command, d.Get("computed_pub_key").(string))
	}
	response := execCmd(command...)
	for _, line := range strings.Split(response.String(), "\n") {
		if keyIdPattern.MatchString(line) {
			keyId = strings.TrimSpace(line)
		}
	}
	if "" == keyId && fromKeyring {
		return fmt.Errorf("could not find pattern '%s' in GPG", searchString)
	}
	err = d.Set("gpg_key_id", keyId)
	return err
}

func exportKeyFiles(d *schema.ResourceData) error {
	keyId := d.Get("gpg_key_id").(string)
	pubKeyFileName := d.Get("computed_pub_key").(string)
	secretKeyFileName := d.Get("computed_secret_key").(string)
	if !simpleFileExists(pubKeyFileName) {
		pubKeyCommand := []string{
			"gpg2",
			"--output",
			pubKeyFileName,
			"--armor",
			"--export",
			keyId,
		}
		pubKeyResponse := execCmd(pubKeyCommand...)
		if !pubKeyResponse.Succeeded() {
			return fmt.Errorf("unable to generate pub key: %s", pubKeyResponse.String())
		}
	}
	if !simpleFileExists(secretKeyFileName) {
		secretKeyCommand := []string{
			"gpg2",
			"--output",
			secretKeyFileName,
			"--armor",
			"--export-secret-key",
			keyId,
		}
		secretKeyResponse := execCmd(secretKeyCommand...)
		if !secretKeyResponse.Succeeded() {
			return fmt.Errorf("unable to generate secret key: %s", secretKeyResponse.String())
		}
	}
	return nil
}

func destroyLocalFiles(d *schema.ResourceData) error {
	files := []string{
		d.Get("computed_pub_key").(string),
		d.Get("computed_secret_key").(string),
	}
	for _, fileName := range files {
		err := os.Remove(fileName)
		if nil != err {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func destroyGpgKeys(d *schema.ResourceData) error {
	keyId := d.Get("gpg_key_id").(string)
	command := []string{
		"gpg2",
		"--batch",
		"--delete-secret-and-public-key",
		"--yes",
		keyId,
	}
	response := execCmd(command...)
	if !response.Succeeded() {
		if 2 != response.exitCode {
			return fmt.Errorf("unable to delete gpg key: %s", response.String())
		}
	}
	return nil
}

func resourceGpgKeysCreate(d *schema.ResourceData, m interface{}) error {
	err := initializeResourceCall(d, false)
	if nil != err {
		return err
	}
	err = execBatchGeneration(d)
	if nil != err {
		return err
	}
	err = determineKeyId(d, true)
	if nil != err {
		return err
	}
	err = exportKeyFiles(d)
	if nil != err {
		return err
	}
	d.SetId(d.Get("gpg_key_id").(string))
	return nil
}

func resourceGpgKeysRead(d *schema.ResourceData, m interface{}) error {
	err := initializeResourceCall(d, true)
	if nil != err {
		return err
	}
	err = determineKeyId(d, false)
	if nil != err {
		return err
	}
	keyId := d.Get("gpg_key_id").(string)
	if "" == keyId {
		err = determineKeyId(d, true)
		if nil != err {
			return err
		}
	}
	return nil
}

func resourceGpgKeysDelete(d *schema.ResourceData, m interface{}) error {
	err := initializeResourceCall(d, true)
	if nil != err {
		return err
	}
	err = determineKeyId(d, false)
	if nil != err {
		return err
	}
	err = destroyLocalFiles(d)
	if nil != err {
		return err
	}
	keyId := d.Get("gpg_key_id").(string)
	if "" == keyId {
		err = determineKeyId(d, true)
		if nil != err {
			return nil
		}
	}
	keyId = d.Get("gpg_key_id").(string)
	if "" != keyId {
		err = destroyGpgKeys(d)
		if nil != err {
			return err
		}
	}
	return nil
}

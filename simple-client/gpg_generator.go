package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

const gpgBatchFile = `
%echo Generating a configuration OpenPGP key
%no-protection
Key-Type: default
Subkey-Type: default
Name-Real: CJ Harries
Name-Comment: Zero Trust Secrets
Name-Email: cj@wotw.pro
Expire-Date: 0
%commit
%echo done
`

var keyIdPattern, _ = regexp.Compile(`^\s+[^\s]*?\s*$`)

func RunGpgBatch() {
	batchFile := WriteToTempFile(gpgBatchFile)
	command := []string{
		"gpg2",
		"--batch",
		"--armor",
		"--gen-key",
		batchFile.Name(),
	}
	response := ExecCmd(command...)
	fmt.Println(response.String())
}

func CheckIfKeyExists() bool {
	command := []string{
		"gpg2",
		"--list-keys",
		"Zero Trust Secrets",
	}
	response := ExecCmd(command...)
	return response.Succeeded()
}

func EnsureKeyExists() {
	if !CheckIfKeyExists() {
		RunGpgBatch()
	}
}

func DetermineKeyId() string {
	command := []string{
		"gpg2",
		"--list-keys",
		"Zero Trust Secrets",
	}
	response := ExecCmd(command...)
	keyId := ""
	for _, line := range strings.Split(response.String(), "\n") {
		if keyIdPattern.MatchString(line) {
			keyId = strings.TrimSpace(line)
		}
	}
	if "" == keyId {
		log.Fatal("No key ID found")
	}
	return keyId
}

func ExportKeyFiles(directory string, keyId string) {
	pubKeyFileName := GetKeyFileName(directory, GpgKeyTypePub)
	if !FileExists(pubKeyFileName) {
		pubKeyCommand := []string{
			"gpg2",
			"--output",
			pubKeyFileName,
			"--armor",
			"--export",
			keyId,
		}
		pubKeyResponse := ExecCmd(pubKeyCommand...)
		if !pubKeyResponse.Succeeded() {
			log.Fatal(pubKeyResponse.exitErr)
		}
	}
	secretKeyFileName := GetKeyFileName(directory, GpgKeyTypeSecret)
	if !FileExists(secretKeyFileName) {
		secretKeyCommand := []string{
			"gpg2",
			"--output",
			secretKeyFileName,
			"--armor",
			"--export-secret-key",
			keyId,
		}
		secretKeyResponse := ExecCmd(secretKeyCommand...)
		if !secretKeyResponse.Succeeded() {
			log.Fatal(secretKeyResponse.exitErr)
		}
	}
}

type GpgKeyType int

const (
	GpgKeyTypePub GpgKeyType = iota
	GpgKeyTypeSecret
)

func GetKeyFileName(directory string, keyType GpgKeyType) string {
	if GpgKeyTypeSecret == keyType {
		return fmt.Sprintf("%s/.secring.gpg", directory)
	}
	return fmt.Sprintf("%s/.pubring.gpg", directory)
}

func EnsureKeyFilesExist(directory string) {
	pubKeyFileName := GetKeyFileName(directory, GpgKeyTypePub)
	secretKeyFileName := GetKeyFileName(directory, GpgKeyTypeSecret)
	if !FileExists(pubKeyFileName) || !FileExists(secretKeyFileName) {
		EnsureKeyExists()
		keyId := DetermineKeyId()
		ExportKeyFiles(directory, keyId)
	}
}

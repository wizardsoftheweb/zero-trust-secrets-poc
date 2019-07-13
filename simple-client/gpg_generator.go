package main

import (
	"fmt"
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

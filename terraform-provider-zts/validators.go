package main

import (
	"fmt"
	"os"
)

func checkIfSchemaStringIsString(schemaString interface{}, stringName string) (string, []error) {
	parsed, ok := schemaString.(string)
	if !ok {
		return "", []error{
			fmt.Errorf("expected %s to be a string", stringName),
		}
	}
	return parsed, []error{}
}

func checkFileObjectExists(objectPath string) (os.FileInfo, []error) {
	stat, err := os.Stat(objectPath)
	if os.IsNotExist(err) {
		return stat, []error{
			fmt.Errorf("%s does not exist", objectPath),
		}
	} else if nil != err {
		return stat, []error{
			fmt.Errorf("unknown io error"),
		}
	}
	return stat, []error{}
}

func validateFileObject(isDirectory bool, objectSchemaString interface{}, k string) ([]string, []error) {
	var warnings []string
	var foundErrors []error
	var objectType string
	if isDirectory {
		objectType = "directory"
	} else {
		objectType = "file"
	}
	parsed, errs := checkIfSchemaStringIsString(objectSchemaString, objectType)
	if 0 < len(errs) {
		foundErrors = append(foundErrors, errs...)
		return warnings, foundErrors
	}
	stat, errs := checkFileObjectExists(parsed)
	if 0 < len(errs) {
		foundErrors = append(foundErrors, errs...)
		return warnings, foundErrors
	}
	if isDirectory && !stat.IsDir() {
		foundErrors = append(foundErrors, fmt.Errorf("%s is not a directory", parsed))
		return warnings, foundErrors
	}
	if !isDirectory && stat.IsDir() {
		foundErrors = append(foundErrors, fmt.Errorf("%s is not a file", parsed))
		return warnings, foundErrors
	}
	return warnings, foundErrors
}

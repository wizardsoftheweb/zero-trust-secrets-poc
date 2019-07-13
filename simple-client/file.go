package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func WriteToTempFile(fileContents string) *os.File {
	file, err := ioutil.TempFile("", "")
	if nil != err {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(
		file.Name(),
		[]byte(fileContents),
		0666,
	)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Println(file.Name())
	return file
}

func FileExists(fileName string) bool {
	stat, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

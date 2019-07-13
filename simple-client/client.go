package main

import "os"

func main() {
	cwd, _ := os.Getwd()
	EnsureKeyFilesExist(cwd)
}

package main

var fatalCheck = func(err error) {
	if nil != err {
		panic(err)
	}
}

func main() {
	BuildServer()
}

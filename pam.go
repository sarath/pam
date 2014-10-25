package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
)

type PamRcConf struct {
	Download string
	Bin string
	Registry string
}

type PamConf struct {

}

var pamrc *PamRcConf

func main() {


	usage := func() {
		fmt.Println("usage pam install|remove pkgname")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		usage()
	}

	initConf()

	cmd := os.Args[1]
	pkg := os.Args[2]

	switch cmd {
	case "install" :
		install(pkg)
	case "remove" :
		remove(pkg)
	default :
		usage();
	}


}


func install(pkg string) {
	fmt.Printf("Installing: %v\n", pkg)
//	pamc := readPamConfig(pkg)
}

func remove(pkg string) {
	fmt.Printf("Removing: %v\n", pkg)
}

func initConf() {
	file, e := ioutil.ReadFile("./config/pam.rc.json") //todo: externalise
	if e != nil {
		pamrc = new(PamRcConf) // todo:defaults
		return
	}
	e = json.Unmarshal(file, &pamrc)
	if e != nil {
		exitFail("Error reading config file content %v\n", e)
	}
	fmt.Printf("bin:%s\n", pamrc.Bin)
}

func exitFail(fm string, a ...interface{}) {
	fmt.Printf(fm,a)
	os.Exit(2)
}



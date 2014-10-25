package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"path"
	"log"
	"net/http"
	"io"
)

type PamRcConf struct {
	Cache    string
	Bin      string
	Registry string
}

//{
//"name": "gitp",
//"version": "1.9.0",
//"description": "Git commands",
//"source": {
//"url": "http://msysgit.googlecode.com/files/PortableGit-1.9.0-preview20140217.7z",
//"type": "7z"
//},
//"path": [
//"bin", "cmd", "share/vim/vim73"
//],
//"env": {}
//}

type PamConf struct {
	Name        string
	Version     string
	Description string
	Source struct {
		Url  string
		Type string
	}
	Path        []string
	Env map[string]string
}

var (
	pamrc *PamRcConf = new(PamRcConf)
	PAM_EXTENSION    = ".pam.json"
	PAM_DEF_RC       = "./config/pam.rc.json"
)

func main() {


	usage := func() {
		log.Fatalf("usage pam install|remove pkgname\n")
	}

	if len(os.Args) < 3 {
		usage()
	}

	initConf()

	cmd := os.Args[1]
	pkg := os.Args[2]

	switch cmd {
	case "install" :
		install(pkg, len(os.Args) >3 && os.Args[3] == "-f")
	case "remove" :
		remove(pkg)
	default :
		usage();
	}


}


func install(pkg string, force bool) {
	log.Printf("Installing: %v, ", pkg)
	var pamc PamConf
	readPamConfig(pkg, &pamc)

	//determine dest
	dest := path.Join(pamrc.Cache, pamc.Name+"-"+pamc.Version+determineType(pamc))
	log.Printf("version: %v, from: %s, to: %s", pamc.Version, pamc.Source.Url, dest)
	err := downloadFile(pamc.Source.Url, dest, force)
	if err != nil {
		log.Fatalf("Error while downloading %v", err)
	}
}

func downloadFile(src string, dest string, force bool) (err error) {

	if _, err := os.Stat(dest); err == nil {
		if (force) {
			log.Println("%s exists, but force set, downloading",dest)
		}else {
			log.Printf("%s exists, skipping download", dest)
			return nil
		}
	}

	out, err := os.Create(dest)
	defer out.Close()
	if err != nil {
		return
	}

	resp, err := http.Get(src)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return
	}
	log.Printf("Download complete [%v bytes]", n)
	return
}

func determineType(pamc PamConf) string {
	if len(pamc.Source.Type) != 0 {
		return "." + pamc.Source.Type
	}
	return path.Ext(pamc.Source.Url)
}

func readPamConfig(pkg string, o *PamConf) {
	p := path.Join(pamrc.Registry, pkg+PAM_EXTENSION)
	e := readJson(p, o)
	if e != nil {
		log.Fatalf("Error reading %v %v", p, e)
	}
}

func remove(pkg string) {
	log.Printf("Removing: %v\n", pkg)
}

func initConf() {
	e := readJson(PAM_DEF_RC, pamrc)
	if e != nil {
		log.Fatalf("Error reading pam.rc.json %v\n", e)
	}
	log.Printf("bin:%s\n", pamrc.Bin)
}

func readJson(filename string, o interface{}) (e error) {
	file, e := ioutil.ReadFile(filename)
	log.Printf("read file succesfully: %s %s", filename, string(file))
	if e == nil {
		e = json.Unmarshal(file, &o)
	}
	return
}




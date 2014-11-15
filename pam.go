package main

import (
	"os"
	"io/ioutil"
	//	"crypto/md5"
	"encoding/json"
	"path"
	"log"
	"net/http"
	"io"
	"os/exec"
	"strings"
	//	"crypto/des"
	"path/filepath"
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
		Url      string
		Type     string
		Checksum []byte
	}
	Path        []string
	Env map[string]string
}

var (
	pamrc *PamRcConf = new(PamRcConf)
	PAM_EXTENSION    = ".pam.json"
	PAMRC            = ".pamrc.json"
	SEVENZA          = "7za"
	PLSEP            = string(os.PathListSeparator)
)

func main() {

	initConf()


	usage := func() {
		log.Fatalf("usage pam install|remove pkgname\n")
	}

	if len(os.Args) < 3 {
		usage()
	}

	cmd := os.Args[1]
	pkg := os.Args[2]

	switch cmd {
	case "install" :
		install(pkg, len(os.Args) > 3 && os.Args[3] == "-f")
	case "remove" :
		remove(pkg)
	default :
		usage();
	}


}


func install(pkg string, force bool) {
	log.Printf("Installing: %v, ", pkg)

	//read and cache config
	var pamc PamConf
	readPamConfig(pkg, &pamc)

	//determine dest
	archFileLoc := path.Join(pamrc.Cache, pamc.Name+"-"+pamc.Version+determineType(pamc))
	log.Printf("version: %v, from: %s, to: %s", pamc.Version, pamc.Source.Url, archFileLoc)

	//download
	err := downloadFile(pamc.Source.Url, archFileLoc, force, pamc.Source.Checksum)
	if err != nil {
		log.Fatal("Error while downloading", err)
	}

	installFolder := getInstallFolder(pkg)
	//extract
	extractArchive(archFileLoc, installFolder, force)

	saveJson(path.Join(installFolder, PAM_EXTENSION), pamc)

	//setup env
	setupEnv(pamc)

	//setup paths
	setupPaths(pamc)

	printComplete(pamc, "installation complete")
}

func remove(pkg string) {
	p := path.Join(pamrc.Bin, pkg)
	pamc := PamConf{}
	readJson(path.Join(p, PAM_EXTENSION), &pamc)
	log.Printf("Removing: %v\n", p)
	os.RemoveAll(p)

	removePath(pamc)

	printComplete(pamc, "removed")
	removeEnv(pamc)
}

func removePath(pamc PamConf) {
	pathv := os.Getenv("PAM_PATH")
	log.Println("Path before:", pathv)
	for _, v := range pamc.Path {
		nv := path.Join(getInstallFolder(pamc.Name), v+PLSEP)
		log.Println("Path :removing:", nv)
		pathv = strings.Replace(pathv, nv, "", -1);
	}
	log.Println("Path after:", pathv)

}

func removeEnv(pamc PamConf) {
	log.Printf("The following Environment variables were setup when installing %v, please remove them manually\n", pamc.Name)
	for k, _ := range pamc.Env {
		log.Print(k, " ")
	}
	launchSysEnv()
}

func launchSysEnv() {
	exeQ([]string{"rundll32", "sysdm.cpl,EditEnvironmentVariables"})
}

func printComplete(pamc PamConf, oplog string) {
	log.Println(pamc.Name, pamc.Version, oplog)
}

func setupPaths(pamc PamConf) {
	if pamc.Path == nil || len(pamc.Path) == 0 {
		log.Println("Setting Path: Nothing to set")
		return
	}
	log.Println("Setting Path :", pamc.Path)
	pamPath := os.Getenv("PAM_PATH")
	for _, v := range pamc.Path {
		nv := path.Join(getInstallFolder(pamc.Name), v+PLSEP)
		if strings.Contains(pamPath, nv) {
			continue
		}
		pamPath = nv+pamPath
		cmds := []string{"setx", "PAM_PATH", pamPath}
		exeQ(cmds)
	}
	setPamPathInPathIfNecessary()
}

func setPamPathInPathIfNecessary() {
	path := os.Getenv("PATH")
	pamPathV := os.Getenv("PAM_PATH") + PLSEP
	if !strings.Contains(path, pamPathV) {
		log.Println("You must add %PAM_PATH% in PATH, once, to make pam installed software run in path")
	}
}

func setupEnv(pamc PamConf) {
	if pamc.Env == nil || len(pamc.Env) == 0 {
		log.Println("Setting Env: Nothing to set")
		return
	}
	log.Println("Setting Env :", pamc.Env)
	for k, v := range pamc.Env {
		nv := strings.Replace(v, "%INSTALL%", getInstallFolder(pamc.Name), -1) //todo test
		setx(k, nv)
	}
}

func setx(k string, v string) {
	cmds := []string{"setx", k, v}
	exeQ(cmds)
}

func getInstallFolder(pkg string) string {
	return path.Join(pamrc.Bin, pkg)
}

func extractArchive(file string, destination string, force bool) {
	if _, err := os.Stat(destination); err == nil && !force {
		log.Println(destination, "exits, user force (-f)")
		return
	}
	log.Println("Extracting:", file)
	sevenza := path.Join(pamrc.Bin, SEVENZA, SEVENZA) //$Bin/7za/7za
	cmds := []string{sevenza, "x", file, "-o" + destination, "-aos"}
	if force {
		cmds[len(cmds)-1] = "-aoa"
	}
	e := exeQ(cmds)
	if e != nil {
		log.Fatal("Error during extraction : ", e)
	}
	log.Printf("Extraction complete")
}

func exeQ(cmds []string) error {
	log.Println("Running:", cmds)
	c := exec.Command(cmds[0], cmds[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func downloadFile(src string, dest string, force bool, checksum []byte) (err error) {

	if _, err := os.Stat(dest); err == nil {
		if (force) {
			log.Println("%s exists, but force set, downloading", dest)
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

	if checksum != nil {
		//		todo: md5.Sum(out)
	}
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

func initConf() {
	pamrcloc := genPamRcLoc()
	log.Println("Reading user's pamrc:", pamrcloc)

	if _, err := os.Stat(pamrcloc); err != nil {
		//create defaults
		// first run
		log.Printf("%s doesnot exist, creating it, setting up first run", pamrcloc)

		pamloc, _ := filepath.Abs(filepath.Dir(os.Args[0]))

		pamrc.Bin = pamloc
		pamrc.Cache = path.Join(pamloc, ".cache")
		pamrc.Registry = path.Join(pamloc, ".reg")

		log.Println("creating", pamrcloc, *pamrc)
		err = saveJson(pamrcloc, pamrc)
		if err != nil {
			log.Fatalf("Unable to create %s", pamrcloc)
		}


		//		os.Mkdir(pamrc.Cache, os.FileMode(644))
		//		os.Mkdir(pamrc.Registry, os.FileMode(644))

		setx("PAM_HOME", pamloc)
		if os.Getenv("PAM_PATH") == "" {
			setx("PAM_PATH", "%PAM_HOME%")
		}
		//		find a way to do this effectively.
		//		if !strings.Contains(os.Getenv("PATH"), "%PAM_PATH%") {
		//			setx("PATH", "%PAM_PATH%;"+os.Getenv("PATH"))
		//		}

		log.Println("Please add %PAM_PATH% to %PATH% manually")
		launchSysEnv()
		os.Exit(0)
	}

	e := readJson(pamrcloc, pamrc)
	if e != nil {
		log.Fatalf("Error reading pamrc.json %v\n", e)
	}
}

func genPamRcLoc() string {
	dest := os.Getenv("HOME")
	if dest == "" {
		dest = os.Getenv("USERPROFILE")
	}

	return path.Join(dest, PAMRC)
}

func readJson(filename string, o interface{}) (e error) {
	file, e := ioutil.ReadFile(filename)
	log.Printf("read file succesfully: %s %s", filename, string(file))
	if e == nil {
		e = json.Unmarshal(file, &o)
	}
	return
}

func saveJson(filename string, o interface{}) (e error) {
	content, _ := json.Marshal(o)
	return ioutil.WriteFile(filename, content, os.FileMode(644))
}



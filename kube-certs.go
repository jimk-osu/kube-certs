/*
  - Create json/yaml config file ingestion.
  - Add configurable cert output param
  - Create logic to check if certs exist already
    for WebHook driven builds when adding new nodes
  - validate certs with ca for n+1 runs
  - Create external accessible cert validation for
    pipeline drive node builds (Ansible, Ignition, etc.)
  - Look into proper cert storage
*/
package main

import (
	//"github.com/cloudflare/cfssl/cmd/cfssl/cfssl"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type KubeConfig struct {
	Node struct {
		Workers []struct {
			Hostname   string `json:"hostname"`
			ExternalIP string `json:"externalIP"`
			InternalIP string `json:"internalIP"`
		} `json:"workers"`
		Controllers []struct {
			Hostname   string `json:"hostname"`
			ExternalIP string `json:"externalIP"`
			InternalIP string `json:"internalIP"`
		} `json:"controllers"`
		KubeApiAddr string `json:"kube-api-addr"`
	} `json:"nodes"`
	CA struct {
		Location string `json:"location"`
		CN       string `json:"CN"`
		ORG      string `json:"ORG"`
		OU       string `json:"OU"`
		ST       string `json:"ST"`
	} `json:"ca"`
}

func main() {
	file, err := ioutil.ReadFile("./kube-cert-config.json")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(file))

	var kubeconfig KubeConfig
	json.Unmarshal(file, &kubeconfig)
	fmt.Printf("Results: %v\n", kubeconfig)

	createCaCSR(kubeconfig)
	genCA()
}

func createCaCSR(kubeconfig KubeConfig) {
	file, err := ioutil.ReadFile("./kube-cert-config.json")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(file))

	// var kubeconfig KubeConfig
	json.Unmarshal(file, &kubeconfig)
	fmt.Printf("Results: %v\n", kubeconfig)

	// Get Variable assingment working
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var caStr = fmt.Sprintf(`
{
	"CN": "Kubernetes",
	"key": {
	"algo": "rsa",
	"size": 2048
	},
	"names": [
		{
			"C": "US",
			"L": "%s",
			"O": "%s",
			"OU": "%s",
			"ST": "%s"
		}
	]
}`, kubeconfig.CA.Location, kubeconfig.CA.ORG, kubeconfig.CA.OU, kubeconfig.CA.ST)
		defer stdin.Close()
		io.WriteString(stdin, caStr)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)

	errWF := ioutil.WriteFile("templateFiles/ca-csr.json", out, 0644)
	check(errWF)
}

func genCA() {
	cmd := exec.Command("cfssl", "gencert", "-initca", "templateFiles/ca-csr.json")
	// stdin, err := cmd.StdinPipe()

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// go func() {

	// 	defer stdin.Close()
	// 	io.WriteString(stdin, `| cfssljson -bare ca`)
	// }()

	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Printf("StdOut")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {

		log.Fatal(err)

	}

	// if err := cmd.Start(); err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Printf("%s\n", stdoutStderr)

	n := bytes.Index(stdoutStderr, []byte{0})

	cmd2 := exec.Command("cfssljson", "-bare", "ca", string(stdoutStderr[:n]))

	err2 := cmd2.Run()
	check(err2)

	log.Printf("Command finished with error: %v", err2)

	// fmt.Printf("%s\n", out)
	// errWF := ioutil.WriteFile("templateFiles/ca-csr.json", out, 0644)
	// check(errWF)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

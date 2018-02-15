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

	//fmt.Printf("%s\n", string(file))

	var kubeconfig KubeConfig
	json.Unmarshal(file, &kubeconfig)
	//fmt.Printf("Results: %v\n", kubeconfig)

	createCaCSR(kubeconfig)
	genCA()
	genAdmin(kubeconfig)
	genWorkers(kubeconfig)
	genKubeProxy(kubeconfig)
	genApiServer(kubeconfig)
}

func createCaCSR(kubeconfig KubeConfig) {
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
	cfssl := exec.Command("cfssl", "gencert", "-initca", "templateFiles/ca-csr.json")

	cfssljson := exec.Command("cfssljson", "-bare", "ca")

	reader, writer := io.Pipe()
	var buf bytes.Buffer

	cfssl.Stdout = writer
	cfssljson.Stdin = reader

	cfssljson.Stdout = &buf

	cfssl.Start()
	cfssljson.Start()

	cfssl.Wait()
	writer.Close()

	cfssljson.Wait()
	reader.Close()

	io.Copy(os.Stdout, &buf)
}

func genAdmin(kubeconfig KubeConfig) {

	// Get Variable assingment working
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var caStr = fmt.Sprintf(`
	{
		"CN": "admin",
		"key": {
			"algo": "rsa",
			"size": 2048
		},
		"names": [
			{
			"C": "US",
			"L": "%s",
			"O": "system:masters",
			"OU": "%s",
			"ST": "%s"
			}
		]
		}`, kubeconfig.CA.Location, kubeconfig.CA.OU, kubeconfig.CA.ST)
		defer stdin.Close()
		io.WriteString(stdin, caStr)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)

	errWF := ioutil.WriteFile("templateFiles/admin-csr.json", out, 0644)
	check(errWF)

	cfssl := exec.Command("cfssl", "gencert", "-ca=ca.pem", "-ca-key=ca-key.pem", "-config=templateFiles/ca-conf.json", "-profile=kubernetes", "templateFiles/admin-csr.json")

	cfssljson := exec.Command("cfssljson", "-bare", "admin")

	reader, writer := io.Pipe()
	var buf bytes.Buffer

	cfssl.Stdout = writer
	cfssljson.Stdin = reader

	cfssljson.Stdout = &buf

	cfssl.Start()
	cfssljson.Start()

	cfssl.Wait()
	writer.Close()

	cfssljson.Wait()
	reader.Close()

	io.Copy(os.Stdout, &buf)

}

func genWorkers(kubeconfig KubeConfig) {

	workerCount := len(kubeconfig.Node.Workers)
	fmt.Printf("Writing %d worker certs\n", workerCount)

	for j := 0; j < workerCount; j++ {
		// var worker  = kubeconfig.Node.Workers[j]
		var instance string
		fmt.Printf("%d\n", j)
		fmt.Printf("External IP:%s\n", kubeconfig.Node.Workers[j].InternalIP)
		if string(kubeconfig.Node.Workers[j].ExternalIP) == "" {
			instance = kubeconfig.Node.Workers[j].InternalIP
		} else {
			instance = kubeconfig.Node.Workers[j].ExternalIP
		}

		// Get Variable assingment working
		cmd := exec.Command("cat")
		stdin, err := cmd.StdinPipe()

		if err != nil {
			log.Fatal(err)
		}

		go func() {
			var caStr = fmt.Sprintf(`
		{
			"CN": "system:node:%s",
			"key": {
				"algo": "rsa",
				"size": 2048
			},
			"names": [
				{
				"C": "US",
				"L": "%s",
				"O": "system:nodes",
				"OU": "%s",
				"ST": "%s"
				}
			]
			}`, instance, kubeconfig.CA.Location, kubeconfig.CA.OU, kubeconfig.CA.ST)
			defer stdin.Close()
			io.WriteString(stdin, caStr)
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", out)

		var workerJsonLocation = "templateFiles/" + instance + "-csr.json"

		errWF := ioutil.WriteFile(workerJsonLocation, out, 0644)
		check(errWF)

		//Add check for optional if exists
		var hostname = "-hostname=" + kubeconfig.Node.Workers[j].InternalIP + "," + kubeconfig.Node.Workers[j].Hostname

		cfssl := exec.Command("cfssl", "gencert", "-ca=ca.pem", "-ca-key=ca-key.pem", "-config=templateFiles/ca-conf.json", hostname, "-profile=kubernetes", "templateFiles/"+kubeconfig.Node.Workers[j].InternalIP+"-csr.json")

		cfssljson := exec.Command("cfssljson", "-bare", kubeconfig.Node.Workers[j].InternalIP)

		reader, writer := io.Pipe()
		var buf bytes.Buffer

		cfssl.Stdout = writer
		cfssljson.Stdin = reader

		cfssljson.Stdout = &buf

		cfssl.Start()
		cfssljson.Start()

		cfssl.Wait()
		writer.Close()

		cfssljson.Wait()
		reader.Close()

		io.Copy(os.Stdout, &buf)
	}

}

func genKubeProxy(kubeconfig KubeConfig) {

	// Get Variable assingment working
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var caStr = fmt.Sprintf(`
	{
		"CN": "system:kube-proxy",
		"key": {
			"algo": "rsa",
			"size": 2048
		},
		"names": [
			{
			"C": "US",
			"L": "%s",
			"O": "system:node-proxier",
			"OU": "%s",
			"ST": "%s"
			}
		]
		}`, kubeconfig.CA.Location, kubeconfig.CA.OU, kubeconfig.CA.ST)
		defer stdin.Close()
		io.WriteString(stdin, caStr)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)

	errWF := ioutil.WriteFile("templateFiles/kube-proxy-csr.json", out, 0644)
	check(errWF)

	cfssl := exec.Command("cfssl", "gencert", "-ca=ca.pem", "-ca-key=ca-key.pem", "-config=templateFiles/ca-conf.json", "-profile=kubernetes", "templateFiles/kube-proxy-csr.json")

	cfssljson := exec.Command("cfssljson", "-bare", "kube-proxy")

	reader, writer := io.Pipe()
	var buf bytes.Buffer

	cfssl.Stdout = writer
	cfssljson.Stdin = reader

	cfssljson.Stdout = &buf

	cfssl.Start()
	cfssljson.Start()

	cfssl.Wait()
	writer.Close()

	cfssljson.Wait()
	reader.Close()

	io.Copy(os.Stdout, &buf)

}

func genApiServer(kubeconfig KubeConfig) {

	// Get Variable assingment working
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var caStr = fmt.Sprintf(`
	{
		"CN": "kubernetes",
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

	errWF := ioutil.WriteFile("templateFiles/kubernetes-csr.json", out, 0644)
	check(errWF)

	var hostname = "-hostname=10.32.0.1,10.240.0.10,10.240.0.11,10.240.0.12," + kubeconfig.Node.KubeApiAddr + ",127.0.0.1,kubernetes.default"

	cfssl := exec.Command("cfssl", "gencert", "-ca=ca.pem", "-ca-key=ca-key.pem", "-config=templateFiles/ca-conf.json", hostname, "-profile=kubernetes", "templateFiles/kubernetes-csr.json")

	cfssljson := exec.Command("cfssljson", "-bare", "kubernetes")

	reader, writer := io.Pipe()
	var buf bytes.Buffer

	cfssl.Stdout = writer
	cfssljson.Stdin = reader

	cfssljson.Stdout = &buf

	cfssl.Start()
	cfssljson.Start()

	cfssl.Wait()
	writer.Close()

	cfssljson.Wait()
	reader.Close()

	io.Copy(os.Stdout, &buf)

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

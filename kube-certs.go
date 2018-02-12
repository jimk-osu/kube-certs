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
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	//"errors"
	"io"
	"io/ioutil"
	"os"
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
	readKubeCertJson()
}

func readKubeCertJson() {
	file, err := ioutil.ReadFile("./kube-cert-config.json")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(file))

	var kubeconfig KubeConfig
	json.Unmarshal(file, &kubeconfig)
	fmt.Printf("Results: %v\n", kubeconfig)

	// Get Variable assingment working
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		// This is broke
		// Create string function to return proper string with verbs assigned
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
		}
		EOF`, kubeconfig.CA.ORG, kubeconfig.CA.Location, kubeconfig.CA.ORG, kubeconfig.CA.OU, kubeconfig.CA.ST)

		defer stdin.Close()
		io.WriteString()
	}()

	/*
		cmd = exec.Command("cat",  `> ca-csr.json <<EOF
		{
		  "CN": "Kubernetes",
		  "key": {
			"algo": "rsa",
			"size": 2048
		  },
		  "names": [
			{
			  "C": "US",
			  "L": "`+"%s"+`",
			  "O": "`+"%s"+`",
			  "OU": "`+"%s"+`",
			  "ST": "`+"%s"+`
			}
		  ]
		}
		EOF`, kubeconfig.CA.ORG, kubeconfig.CA.Location, kubeconfig.CA.ORG, kubeconfig.CA.OU, kubeconfig.CA.ST)

	*/

}

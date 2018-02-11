package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var kubeConf = []byte(`{
	"nodes" : {
	  "workers" : [
		{ "hostname" : "kubeworker3",
		  "externalIP": "",
		  "internalIP": "192.168.56.109"
		},
		{ "hostname" : "kubeworker4",
		  "externalIP": "",
		  "internalIP": "192.168.56.110"
		}
	  ],
	  "controllers": [
		{ "hostname" : "kubemaster2",
		  "externalIP": "",
		  "internalIP": "192.168.56.108"
		}
	  ],
	  "kube-api-addr": "192.168.56.108"
	},
	"ca" : {
	  "location": "Columbus",
	  "CN": "Kube-Certs",
	  "ORG": "Acme",
	  "OU": "OGs",
	  "ST": "Ohio"
	}
  }
  `)

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

	//var c []Page
	var jsontype KubeConfig
	json.Unmarshal(file, &jsontype)
	fmt.Printf("Results: %v\n", jsontype)

	fmt.Println(jsontype.Node.KubeApiAddr)
}

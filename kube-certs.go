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
  "fmt"
  "encoding/json"
  //"errors"
  "io/ioutil"
  "os")

// type Worker struct {
//   node Kubenode[]
// }
//
// type Controller struct {
//   node Kubenode[]
// }

type KubecertConfg struct{
 Nodes struct {
  Workers []struct {
    Hostname string'json:"hostname"'
    ExternalIP string 'json:"externalIP"'
    InternalIP string 'json:"internalIP"'
  } `json:"workers"`
  Controllers []struct {
    Hostname string 'json:"hostname"'
    ExternalIP string 'json:"externalIP"'
    InternalIP string 'json:"internalIP"'
  } `json:"controllers"`
 } 'json:"nodes"'
 CA struct {
   location string 'json:"location"'
   CN string 'json:"CN"'
   ORG string 'json:"ORG"'
   OU string 'json:"OU"'
   ST string 'json:"ST"'
 } 'json:"ca"'
}

func main(){
  readKubeCertJson()
}

func readKubeCertJson(){
  file, err := ioutil.ReadFile("./kube-cert-config.json")

  if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    fmt.Printf("%s\n", string(file))

    //var c []Page
    //var jsontype Kubecertconfig
    //json.Unmarshal(file, &jsontype)
    // fmt.Printf("Results: %v\n", jsontype)
    var f interface{}
    err := json.Unmarshal(file, $f)

    fmt.Printf("Results: %v\n", f)
}

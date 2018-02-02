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

import ("github.com/cloudflare/cfssl/cmd/cfssl/cfssl"
  "fmt"
  "encoding/json"
  "errors"
  "io/ioutil"
  "os")



func main(){
  fmt.Println()
}

func readKubeCertJson(){
  raw, err := ioutil.ReadFile("./kube-cert.json")

  if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    var c []Page
    json.Unmarshal(raw, &c)
    return c
}

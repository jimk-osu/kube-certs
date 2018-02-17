# kube-certs
A kube certificate chain generator that leverages cfssl. Really a first attempt at Go. Bear with me.

Pre-Reqs
- [cfssl](https://github.com/cloudflare/cfssl) already installed

### Usage
Modify the `kube-certs-config.json` file to reflect the your cluster's internal/external IPs, api servier endpoint, and CA attributes

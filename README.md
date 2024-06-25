<p align="center" style="margin-top: 120px">

  <h3 align="center">homer-k8s</h3>

  <p align="center">
  <img src="https://github.com/BananaOps/homer-k8s/blob/main/images/banner.png?raw=true" style="width:66%" alt="homer-k8s">
  </p>
  <p align="center">
    An Open-Source kubernetes controller to use Homer on k8s
    <br />
  </p>
</p>
<p align="center">
  <a href="https://github.com/BananaOps/homer-k8s/releases"><img title="Release" src="https://img.shields.io/github/v/release/BananaOps/homer-k8s"/></a>
  <a href=""><img title="Downloads" src="https://img.shields.io/github/downloads/BananaOps/homer-k8s/total.svg"/></a>
  <a href=""><img title="Docker pulls" src="https://img.shields.io/docker/pulls/bananaops/homer-k8s"/></a>
  <a href=""><img title="Go version" src="https://img.shields.io/github/go-mod/go-version/BananaOps/homer-k8s"/></a>
  <a href=""><img title="Docker builds" src="https://img.shields.io/docker/automated/bananaops/homer-k8s"/></a>
  <a href=""><img title="Code builds" src="https://img.shields.io/github/actions/workflow/status/BananaOps/homer-k8s/release.yml"/></a>
  <a href=""><img title="apache licence" src="https://img.shields.io/badge/License-Apache-yellow.svg"/></a>
  <a href="https://github.com/BananaOps/homer-k8s/releases"><img title="Release date" src="https://img.shields.io/github/release-date/BananaOps/homer-k8s"/></a>
</p>



## About Homer-k8s 

Homer is a dead simple static HOMepage for your servER to keep your services on hand, from a simple yaml configuration file.

Homer Dashboard https://github.com/bastienwirtz/homer

This project in to order to facilite the deployment in Kubernetes Cluster with dynamic CRDs to define Service aand reload the configuration on each CRDs change.

## Features

- [x] CRDs to define services
- [x] Helm Chart to deploy homer
- [x] Manage homer config in helm values

## Getting Started 🚀

### Requirements

- [golang](https://go.dev/) >= 1.22
- [ko-build](https://ko.build/)
- [helm](https://helm.sh/)


### Crds HomerServices

```yaml
apiVersion: homer.bananaops.io/v1alpha1
kind: HomerServices
metadata:
  labels:
    app.kubernetes.io/name: homer-k8s
  name: homerservices-sample
spec:
  # page: Page 1 # if empty, goes to main page
  groups:
    - name: ci
      icon: "fas fa-code-branch"
      items:
        - name: "Awesome app"
          logo: "assets/tools/sample.png"
          tagstyle: "is-success"
          icon: "fab fa-jenkins"
          subtitle: "Bookmark example"
          tag: "app"
          keywords: "self hosted reddit" # optional keyword used for searching purpose
          url: "https://www.reddit.com/r/selfhosted/"
          target: "_blank" # optional html tag target attribute

          # background: red # optional color for card to set color directly without custom stylesheet
        - name: "Another one"
          logo: "assets/tools/sample.png"
          subtitle: "Another application"
          tag: "app"
          # Optional tagstyle
          tagstyle: "is-success"
          url: "#"
```


### Build 

To compile homer-k8s run this command, output a binnary in bin/event

```bash
skaffold build
```

### Update Manifest

To updates manifest files : 

```bash
make manifest
```

### Deploy with skaffold

To deploy with skaffold: 

! Need to modify skafflod config

```bash
skaffold run
```

### Deploy with Helm


```bash
```sh
helm repo add bananaops https://bananaops.github.io/homer-k8s/
helm repo update bananaops

# install with all defaults
helm install homer bananaops/homer-k8s

# install with customisations
wget https://raw.githubusercontent.com/bananaops/homer-k8s/main/helm/homer-k8s/values.yaml
# edit values.yaml
helm install homer bananaops/homer-k8s -f values.yaml
```


## Contributing

Please see the [contribution guidelines](https://github.com/BananaOps/homer-k8s/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/BananaOps/homer-k8s/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [Apache 2.0 open source license](https://github.com/BananaOps/homer-k8s/blob/main/LICENSE).

`help wanted` issues:
- [homer-k8s](https://github.com/BananaOps/homer-k8s/labels/help%20wanted)

# AALM CICD
The continuous integration and continuous delivery of the Altemista Asset Lifecycle Manager is based on 3 GitHub Actions:

- `build-and-push-image.yaml`
- `process-template.yaml`
- `release.yaml`

## Build and Push Image
This GitHub action is triggered when a change in the code is committed and uses [`operator-sdk`](https://github.com/operator-framework/operator-sdk) to build the operator and docker client to push the built image.

## Process Template
This GitHub action is triggered when any value is changed in `deploy/templates/` directory and will process, with `HELM`, the Harbor and KubeApps Charts to create the following manifests:

- KubeApps:
  - `kubeapps.yaml`: main manifest to install all the KubeApps resources (`Roles`, `Deployment`, `Namespace`, ...)
  - `kubeapps-apprepositories.yaml`: configure AppRepositories in KubeApps to point to Public Catalog and private Harbor
  - `kubeapps-tests.yaml`: manifest to test the installation of KubeApps
  - `kubeapps-cleanup.yaml`: manifest to cleanup some of the KubeApps resources
- Harbor:
  - `harbor.yaml`: main manifest to install Harbor resources
  - `harbor-tests.yaml`: manifest to test the installation of Harbor

## Release
When a new tag is created, a new release version it will be created with this GitHub Action with all the resources needed for the AALM installation.
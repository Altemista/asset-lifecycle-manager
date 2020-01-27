# Altemista Asset Lifecycle Manager
This project is a component of the *Altemista Asset Catalog*, an open source toolkit to manage Kubernetes self-operated applications provided by our Asset Catalog.

AALM extends OLM to provide a declarative way to install, manage, and upgrade assets or applications managed by their Operators in a cluster.

It also enforces some constraints on the components it manages in order to ensure a good user experience.

All the components are installed in the `altemistahub` namespace.

## Installation
Follow the steps in the [installation guide](docs/install-alm.md) to learn how to install the Altemista Asset Lifecycle Manager tool. 

## Components

### OLM
Operator lifecycle manager to manage new installed operators from Operator hub. It is always installed to be able to use the OLM CRs and operators.

More information can be found [here](https://github.com/operator-framework/operator-lifecycle-manager).

### AALM
Altemista OLM to manage Altemista operators. This image it is installed via `CatalogSource`. It is always installed because contains all the information to install and manage Altemista Assets.

### KUBEAPPS
UI that shows all Altemista Assets (when installed) and (optional) harbor private assets and allows to install those assets in the private catalog.

By default, it it automatically synchronized with Altemista public catalog Assets but this funcionality can be disabled with `--no-use-altemista-public-catalog` argument.

More information can be found [here](https://github.com/kubeapps/kubeapps).

### HARBOR
It provides different funcionalities but it is used as Chart repository to store private HELM Charts that can be installed from Kubeapps. When installed, Kubeapps automatically shows the HELM charts from this Chart repository.

More information can be found [here](https://github.com/goharbor/harbor).
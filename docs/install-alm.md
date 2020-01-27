# Install the Altemista Lifecycle Manager
## Usage
This command uses your `KUBECONFIG` environment variable to connect to your Kubernetes. You can configure your installation with the following options:

```bash
Usage: [--(no-)with-kubeapps] [--kubeapps-hostname <arg>] [--(no-)with-harbor] [--harbor-hostname <arg>] [--(no-)use-altemista-public-catalog] [-V|--verbose] [-v|--version] [-h|--help]
    --with-kubeapps, --no-with-kubeapps: install a configured kubeapps marketplace which sync assets from altemista public catalog by default (off by default)
    --kubeapps-hostname: kubeapps hostname (this parameter is required if 'with-kubeapps' flag is marked) (no default)
    --with-harbor, --no-with-harbor: install a configured harbor registry to allow publish own private assets in kubeapps marketplace (off by default)
    --harbor-hostname: harbor hostname (this parameter is required if 'with-harbor' flag is marked) (no default)
    --use-altemista-public-catalog, --no-use-altemista-public-catalog: sync altemista public catalog with kubeapps marketplace (on by default)
    -V, --verbose: verbosity
    -v, --version: Prints version
    -h, --help: Prints help
```

## Install from GitHub release 
This command will install the Altemista Lifecycle Manager with the default arguments.
```bash
curl -fsSL https://github.com/Altemista/asset-lifecycle-manager/releases/latest/download/install.sh | sh
```

If you want to provide arguments with the previous command you can provide arguments to the `sh` command as follows:
```bash
curl -fsSL https://github.com/Altemista/asset-lifecycle-manager/releases/latest/download/install.sh | sh -s -- --with-kubeapps --kubeapps-hostname my-kubeapps.my-company.com --with-harbor --harbor-hostname my-harbor.my-company.com
```
#!/usr/bin/env bash

set -e
set -u

url=https://github.com/Altemista/asset-lifecycle-manager/releases/latest/download
version=$(curl -fsSL https://api.github.com/repos/Altemista/asset-lifecycle-manager/releases/latest \
  | grep tag_name | cut -d : -f 2 | tr -d \",)

# THE DEFAULTS INITIALIZATION - OPTIONALS
_arg_with_kubeapps="off"
_arg_kubeapps_hostname=
_arg_with_harbor="off"
_arg_harbor_hostname=
_arg_use_altemista_public_catalog="on"
_arg_verbose=0

die() {
	local _ret=$2
	test -n "$_ret" || _ret=1
	test "$_PRINT_HELP" = yes && print_help >&2
	echo "$1" >&2
	exit ${_ret}
}

begins_with_short_option() {
	local first_option all_short_options='Vvh'
	first_option="${1:0:1}"
	test "$all_short_options" = "${all_short_options/$first_option/}" && return 1 || return 0
}

print_help() {
	printf '%s\n' "AALM installer help"
	printf 'Usage: %s [--(no-)with-kubeapps] [--kubeapps-hostname <arg>] [--(no-)with-harbor] [--harbor-hostname <arg>] [--(no-)use-altemista-public-catalog] [-V|--verbose] [-v|--version] [-h|--help]\n' "$0"
	printf '\t%s\n' "--with-kubeapps, --no-with-kubeapps: install a configured kubeapps marketplace which sync assets from altemista public catalog by default (off by default)"
	printf '\t%s\n' "--kubeapps-hostname: kubeapps hostname (this parameter is required if 'with-kubeapps' flag is marked) (no default)"
	printf '\t%s\n' "--with-harbor, --no-with-harbor: install a configured harbor registry to allow publish own private assets in kubeapps marketplace (off by default)"
	printf '\t%s\n' "--harbor-hostname: harbor hostname (this parameter is required if 'with-harbor' flag is marked) (no default)"
	printf '\t%s\n' "--use-altemista-public-catalog, --no-use-altemista-public-catalog: sync altemista public catalog with kubeapps marketplace (on by default)"
	printf '\t%s\n' "-V, --verbose: verbosity"
	printf '\t%s\n' "-v, --version: Prints version"
	printf '\t%s\n' "-h, --help: Prints help"
}

parse_commandline() {
	while test $# -gt 0
	do
		_key="$1"
		case "$_key" in
			# The with-kubeapps argurment doesn't accept a value,
			# we expect the --with-kubeapps, so we watch for it.
			--no-with-kubeapps|--with-kubeapps)
				_arg_with_kubeapps="on"
				test "${1:0:5}" = "--no-" && _arg_with_kubeapps="off"
				;;
			# We support whitespace as a delimiter between option argument and its value.
			# Therefore, we expect the --kubeapps-hostname value, so we watch for --kubeapps-hostname.
			# Since we know that we got the long option,
			# we just reach out for the next argument to get the value.
			--kubeapps-hostname)
				test $# -lt 2 && die "Missing value for the optional argument '$_key'." 1
				_arg_kubeapps_hostname="$2"
				shift
				;;
			# We support the = as a delimiter between option argument and its value.
			# Therefore, we expect --kubeapps-hostname=value, so we watch for --kubeapps-hostname=*
			# For whatever we get, we strip '--kubeapps-hostname=' using the ${var##--kubeapps-hostname=} notation
			# to get the argument value
			--kubeapps-hostname=*)
				_arg_kubeapps_hostname="${_key##--kubeapps-hostname=}"
				;;
			# See the comment of option '--with-kubeapps' to see what's going on here - principle is the same.
			--no-with-harbor|--with-harbor)
				_arg_with_harbor="on"
				test "${1:0:5}" = "--no-" && _arg_with_harbor="off"
				;;
			# See the comment of option '--kubeapps-hostname' to see what's going on here - principle is the same.
			--harbor-hostname)
				test $# -lt 2 && die "Missing value for the optional argument '$_key'." 1
				_arg_harbor_hostname="$2"
				shift
				;;
			# See the comment of option '--kubeapps-hostname=' to see what's going on here - principle is the same.
			--harbor-hostname=*)
				_arg_harbor_hostname="${_key##--harbor-hostname=}"
				;;
			# See the comment of option '--with-kubeapps' to see what's going on here - principle is the same.
			--no-use-altemista-public-catalog|--use-altemista-public-catalog)
				_arg_use_altemista_public_catalog="on"
				test "${1:0:5}" = "--no-" && _arg_use_altemista_public_catalog="off"
				;;
			# See the comment of option '--with-kubeapps' to see what's going on here - principle is the same.
			-V|--verbose)
				_arg_verbose=$((_arg_verbose + 1))
				;;
			# We support getopts-style short arguments clustering,
			# so as -V doesn't accept value, other short options may be appended to it, so we watch for -V*.
			# After stripping the leading -V from the argument, we have to make sure
			# that the first character that follows coresponds to a short option.
			-V*)
				_arg_verbose=$((_arg_verbose + 1))
				_next="${_key##-V}"
				if test -n "$_next" -a "$_next" != "$_key"
				then
					{ begins_with_short_option "$_next" && shift && set -- "-V" "-${_next}" "$@"; } || die "The short option '$_key' can't be decomposed to ${_key:0:2} and -${_key:2}, because ${_key:0:2} doesn't accept value and '-${_key:2:1}' doesn't correspond to a short option."
				fi
				;;
			# See the comment of option '--with-kubeapps' to see what's going on here - principle is the same.
			-v|--version)
				echo test v$version
				exit 0
				;;
			# See the comment of option '-V' to see what's going on here - principle is the same.
			-v*)
				echo test v$version
				exit 0
				;;
			# See the comment of option '--with-kubeapps' to see what's going on here - principle is the same.
			-h|--help)
				print_help
				exit 0
				;;
			# See the comment of option '-V' to see what's going on here - principle is the same.
			-h*)
				print_help
				exit 0
				;;
			*)
				_PRINT_HELP=yes die "FATAL ERROR: Got an unexpected argument '$1'" 1
				;;
		esac
		shift
	done
}

install_kubeapps() {
  echo "Installing kubeapps..."
  mongodb_password="$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)"
  curl -fsSL ${url}/kubeapps.yaml \
  | sed -e "s|\${KUBEAPPS_HOSTNAME}|${_arg_kubeapps_hostname}|g" \
        -e "s/mongodb-root-password: \".*\"/mongodb-root-password: \"${mongodb_password}\"/g" \
  | kubectl apply -n altemistahub -f -
}

install_harbor() {
  echo "Installing harbor..."
  echo "Not yet."
}

install_olm() {
  echo "Installing olm..."
  curl -fsSL https://github.com/operator-framework/operator-lifecycle-manager/releases/latest/download/install.sh | bash -s 0.12.0
}

install_altemista_operator_registry() {
  echo "Installing altemista operator registry..."
  curl -fsSL ${url}/aolm.yaml | kubectl apply -f -
}

install() {
  parse_commandline "$@"
  install_olm
  install_altemista_operator_registry
  if [ "$_arg_with_kubeapps" == "on" ]; then
    install_kubeapps
  fi
  if [ "$_arg_with_harbor" == "on" ]; then
    install_harbor
  fi
}

install "$@"

echo "Value of --with-kubeapps is $_arg_with_kubeapps"
echo "Value of --kubeapps-hostname is $_arg_kubeapps_hostname"
echo "Value of --with-harbor is $_arg_with_harbor"
echo "Value of --harbor-hostname is $_arg_harbor_hostname"
echo "Value of --use-altemista-public-catalog is $_arg_use_altemista_public_catalog"
echo "Verbosity degree: $_arg_verbose"

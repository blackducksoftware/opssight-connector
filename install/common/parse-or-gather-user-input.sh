#!/bin/bash
source `dirname ${BASH_SOURCE}`/args.sh "${@}"

function prompt() {
  if [[ $_arg_prompt == "on" ]]; then
    clear
    echo "============================================"
    echo "Blackduck Hub Configuration Information"
    echo "============================================"
    echo "Interactive"
    echo "============================================"
    read -p "Hub server host (e.g. hub.mydomain.com): " hub_host
    read -p "Hub server port [8443]: " hub_port
    read -p "Hub user name [sysadmin]: " hub_user
    read -sp "Hub user password : " _arg_hub_password
    echo " "
    read -p "Maximum concurrent scans [7]: " hub_max_concurrent_scans
    echo "============================================"
  else
    echo "Skipping prompts, --prompt was turned off."
  fi
}

prompt

_arg_hub_host=${hub_host:-$_arg_hub_host}
_arg_hub_port=${hub_port:-$_arg_hub_port}
_arg_hub_user=${hub_user:-$_arg_hub_user}
_arg_hub_max_concurrent_scans=${hub_max_concurrent_scans:-$_arg_hub_max_concurrent_scans}

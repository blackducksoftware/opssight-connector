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
    read -p "Hub server host (e.g. hub.mydomain.com): " _arg_hub_host
    read -p "Hub server port [443]: " _arg_hub_port
    read -p "Hub user name [sysadmin]: " _arg_hub_user
    read -sp "Hub user password : " _arg_hub_password
    echo " "
    read -p "Maximum concurrent scans [7]: " _arg_hub_max_concurrent_scans
    echo "============================================"
  else
    echo "Skipping prompts, --proto_prompty was turned off."
  fi
}

prompt

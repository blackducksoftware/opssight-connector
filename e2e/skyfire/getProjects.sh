#!/bin/bash
  
HOSTNAME=aci-471-aci-471.10.1.176.130.xip.io

BEARER_TOKEN=$(curl -si --insecure --header "Content-Type: application/x-www-form-urlencoded" --request POST --data 'j_username=sysadmin&j_password=blackduck' https://$HOSTNAME:443/j_spring_security_check | awk '/Set-Cookie/{print $2}' | sed 's/;$//')

TOTAL_PROJECT_COUNT=$(curl --insecure -sX GET -H "Accept: application/json" -H "Cookie: $BEARER_TOKEN" https://$HOSTNAME:443/api/projects?limit=0 | jq .totalCount)

echo $TOTAL_PROJECT_COUNT

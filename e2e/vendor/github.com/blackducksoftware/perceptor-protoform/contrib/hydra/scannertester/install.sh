#!/bin/bash

set +x
NS=bds-perceptor

cleanup() {
	oc delete ns $NS
	while oc get ns | grep -q $NS ; do
	  echo "Waiting for deletion...`oc get ns | grep $NS` "
	  sleep 1
	done
}

cleanup

oc new-project $NS

cat << EOF > aux-config.json
{
	"Namespace": "$NS",
	"IsOpenshift": true
}
EOF

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./scannertester.go ./config.json aux-config.json

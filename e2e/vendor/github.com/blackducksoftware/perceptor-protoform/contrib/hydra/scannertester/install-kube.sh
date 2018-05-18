#!/bin/bash

set +x
NS=bds-perceptor


kubectl delete ns $NS
while kubectl get ns | grep -q $NS ; do
  echo "Waiting for deletion...`kubectl get ns | grep $NS` "
  sleep 1
done

cat << EOF > aux-config.json
{
  "Namespace": "$NS",
  "IsOpenshift": false
}
EOF

kubectl create ns $NS

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./scannertester.go ./config.json aux-config.json

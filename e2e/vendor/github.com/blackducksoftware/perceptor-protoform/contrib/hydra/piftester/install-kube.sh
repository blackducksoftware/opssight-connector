#!/bin/bash

set +x
NS=bds-perceptor

IMAGEFACADE_SA="imagefacade-sa"
IMAGE_PERCEIVER_SA="image-perceiver-sa"
POD_PERCEIVER_SA="pod-perceiver-sa"


kubectl delete ns $NS
while kubectl get ns | grep -q $NS ; do
  echo "Waiting for deletion...`kubectl get ns | grep $NS` "
  sleep 1
done


kubectl create ns $NS
kubectl create sa $IMAGEFACADE_SA -n $NS
kubectl create sa $POD_PERCEIVER_SA -n $NS


## finished initial setup, now run piftester

cat << EOF > aux-config.json
{
	"Namespace": "$NS",
	"DockerUsername": "TODO -- REMOVE",
	"DockerPassword": "TODO -- REMOVE",
	"PodPerceiverServiceAccountName": "$POD_PERCEIVER_SA",
	"ImagePerceiverServiceAccountName": "TODO -- REMOVE",
	"ImageFacadeServiceAccountName": "$IMAGEFACADE_SA",
	"InternalDockerRegistries": []
}
EOF

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./piftester.go ./config.json aux-config.json

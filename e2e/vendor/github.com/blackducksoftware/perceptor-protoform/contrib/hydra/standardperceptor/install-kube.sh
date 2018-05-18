#!/bin/bash

set +x
NS=bds-perceptor


IMAGEFACADE_SA="imagefacade-sa"
POD_PERCEIVER_SA="pod-perceiver-sa"


kubectl delete ns $NS
while kubectl get ns | grep -q $NS ; do
  echo "Waiting for deletion...`kubectl get ns | grep $NS` "
  sleep 1
done


kubectl create ns $NS
kubectl create sa $IMAGEFACADE_SA -n $NS
kubectl create sa $POD_PERCEIVER_SA -n $NS


## finished initial setup, now run protoform

cat << EOF > aux-config.json
{
	"Namespace": "$NS",
	"DockerUsername": "TODO -- REMOVE",
	"DockerPassword": "TODO -- REMOVE",
	"PodPerceiverServiceAccountName": "$POD_PERCEIVER_SA",
	"ImagePerceiverServiceAccountName": "TODO -- REMOVE",
	"ImageFacadeServiceAccountName": "$IMAGEFACADE_SA",
	"InternalDockerRegistries": [],
	"IsKube": true
}
EOF

kubectl create -f ../prometheus-deployment.yaml --namespace=$NS

go run ./perceptor.go ./config.json aux-config.json

# If running on minikube:
#   minikube service prometheus -n bds-perceptor --url
#   kubectl expose service perceptor --port=3001 --type=NodePort --name=perceptor-3 -n bds-perceptor
#   kubectl expose service skyfire --port=3187 --type=NodePort --name=skyfire-4 -n bds-perceptor
#
# Otherwise:
#   kubectl expose service prometheus --name=prometheus-metrics --port=9090 --target-port=9090 --type=LoadBalancer -n bds-perceptor

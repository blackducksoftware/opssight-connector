#!/bin/bash

CPUS=$1

cat << EOF > create-hub.yml
apiVersion: synopsys.com/v1
kind: Hub
metadata:
  clusterName: ""
  creationTimestamp: 2018-10-16T18:42:02Z
  generation: 1
  name: aci-471
  namespace: ""
  resourceVersion: "3281480"
  selfLink: /apis/synopsys.com/v1/hubs/aci-471
  uid: 2c53abdc-d173-11e8-b49a-005056b9215d
spec:
  backupInterval: ""
  backupSupport: "No"
  backupUnit: ""
  certificate: ""
  certificateKey: ""
  certificateName: default
  dbPrototype: empty
  dockerRegistry: docker.io
  dockerRepo: blackducksoftware
  environs: []
  flavor: medium
  hubType: worker
  hubVersion: 4.7.1
  imagePrefix: hub
  imageTagMap: null
  namespace: aci-471
  nfsServer: ""
  pvcClaimSize: ""
  pvcStorageClass: ""
  scanType: ""
  state: running
EOF
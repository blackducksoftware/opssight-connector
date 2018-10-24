#!/bin/bash

CPUS=$1

cat << EOF > hub.yml
apiVersion: synopsys.com/v1
kind: Hub
metadata:
  name: smoke-test
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
  nfsServer: ""
  pvcClaimSize: ""
  pvcStorageClass: ""
  scanType: ""
EOF
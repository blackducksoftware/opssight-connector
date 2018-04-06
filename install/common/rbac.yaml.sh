#!/bin/bash
cat << EOF > rbac.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: protoform
subjects:
- kind: ServiceAccount
  name: protoform
  namespace: ${_arg_pcp_namespace}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: perceiver
  namespace: ${_arg_pcp_namespace}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: perceptor-scanner-sa
  namespace: ${_arg_pcp_namespace}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: protoform
  namespace: ${_arg_pcp_namespace}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: all-permission-on-pod-for-perceiver
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "watch", "list","update"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: perceiver
subjects:
- kind: ServiceAccount
  name: perceiver
  namespace: ${_arg_pcp_namespace}
roleRef:
  kind: ClusterRole
  name: all-permission-on-pod-for-perceiver
  apiGroup: rbac.authorization.k8s.io
EOF

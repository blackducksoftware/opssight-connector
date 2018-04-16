In order to deploy on openshift:

- You need to give cluster-admin credentials to the service account, so it can list, and update objects in the existing openshift namespaces.

To do this, run:

- `oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:<project>:openshift-perceiver`, for a baseline of functionality.

- Also, you may need to run `oc adm policy add-scc-to-user privileged system:serviceaccount:<project>:openshift-perceiver` in order to be able to see and update all objects in your cluster.

package kubebuilder

import (
	"k8s.io/api/core/v1"
)

type Resources interface {
	GetConfigMaps() []*v1.ConfigMap
	GetServices() []*v1.Service
	GetSecrets() []*v1.Secret
	GetReplicationControllers() []*v1.ReplicationController
}

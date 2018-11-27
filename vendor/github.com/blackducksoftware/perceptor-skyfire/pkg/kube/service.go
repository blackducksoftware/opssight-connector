package kube

type Service struct {
	Name      string
	Namespace string
	Ports     []int32
}

func NewService(name string, namespace string, port []int32) *Service {
	return &Service{Name: name, Namespace: namespace, Ports: port}
}

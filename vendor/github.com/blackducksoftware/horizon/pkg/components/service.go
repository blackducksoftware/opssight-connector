/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"
	"github.com/koki/short/util/intbool"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Service defines the service component
type Service struct {
	obj *types.Service
}

// NewService creates a Service object
func NewService(config api.ServiceConfig) *Service {
	var affinity *intbool.IntOrBool

	if boolVal, err := strconv.ParseBool(config.Affinity); err != nil {
		affinity = intbool.FromBool(boolVal)
	} else if intVal, err := strconv.Atoi(config.Affinity); err != nil {
		affinity = intbool.FromInt(intVal)
	}

	s := &types.Service{
		Version:                  config.APIVersion,
		Cluster:                  config.ClusterName,
		Name:                     config.Name,
		Namespace:                config.Namespace,
		ClientIPAffinity:         affinity,
		ExternalName:             config.ExternalName,
		ClusterIP:                types.ClusterIP(config.ClusterIP),
		PublishNotReadyAddresses: config.PublishNotReadyAddresses,
	}

	switch config.IPServiceType {
	case api.ClusterIPServiceTypeDefault:
		s.Type = types.ClusterIPServiceTypeDefault
	case api.ClusterIPServiceTypeNodePort:
		s.Type = types.ClusterIPServiceTypeNodePort
	case api.ClusterIPServiceTypeLoadBalancer:
		s.Type = types.ClusterIPServiceTypeLoadBalancer
	}

	switch config.TrafficPolicy {
	case api.ServiceTrafficPolicyNil:
		s.ExternalTrafficPolicy = types.ExternalTrafficPolicyNil
	case api.ServiceTrafficPolicyLocal:
		s.ExternalTrafficPolicy = types.ExternalTrafficPolicyLocal
	case api.ServiceTrafficPolicyCluster:
		s.ExternalTrafficPolicy = types.ExternalTrafficPolicyCluster
	}

	return &Service{obj: s}
}

// GetObj returns the service object in a format the deployer can use
func (s *Service) GetObj() *types.Service {
	return s.obj
}

// GetName returns the name of the service
func (s *Service) GetName() string {
	return s.obj.Name
}

// AddAnnotations adds annotations to the service
func (s *Service) AddAnnotations(new map[string]string) {
	s.obj.Annotations = util.MapMerge(s.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the service
func (s *Service) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		s.obj.Annotations = util.RemoveElement(s.obj.Annotations, k)
	}
}

// AddLabels add labels to the service
func (s *Service) AddLabels(new map[string]string) {
	s.obj.Labels = util.MapMerge(s.obj.Labels, new)
}

// RemoveLabels removes labels from the service
func (s *Service) RemoveLabels(remove []string) {
	for _, k := range remove {
		s.obj.Labels = util.RemoveElement(s.obj.Labels, k)
	}
}

// AddSelectors adds selectors to the service
func (s *Service) AddSelectors(new map[string]string) {
	s.obj.Selector = util.MapMerge(s.obj.Selector, new)
}

// RemoveSelectors removes selectors from the service
func (s *Service) RemoveSelectors(remove map[string]string) {
	for _, k := range remove {
		s.obj.Selector = util.RemoveElement(s.obj.Selector, k)
	}
}

// AddPort adds a port to the service.  There can only be 1 non-named
// port on the service.  If there are more than 1 port then they must all
// be named.  If an unnamed port is already added to the service and another
// port is attmepted to be added, this function will return an error
func (s *Service) AddPort(config api.ServicePortConfig) error {
	if len(config.Name) > 0 && (s.obj.Port != nil || s.obj.NodePort != 0) {
		return fmt.Errorf("all ports must be named if there is more than 1 port")
	}

	port := intstr.Parse(config.TargetPort)
	sp := types.ServicePort{
		Expose:  config.Port,
		PodPort: &port,
	}

	switch config.Protocol {
	case api.ProtocolTCP:
		sp.Protocol = types.ProtocolTCP
	case api.ProtocolUDP:
		sp.Protocol = types.ProtocolUDP
	}

	if len(config.Name) > 0 {
		nsp := types.NamedServicePort{
			Name:     config.Name,
			Port:     sp,
			NodePort: config.NodePort,
		}
		s.obj.Ports = append(s.obj.Ports, nsp)
	} else {
		// Unnamed port
		s.obj.Port = &sp
		s.obj.NodePort = config.NodePort
	}

	return nil
}

// RemovePort will remove a port from the service
func (s *Service) RemovePort(port int32) error {
	if s.obj.Port.Expose == port {
		s.obj.Port = nil
		s.obj.NodePort = 0
		return nil
	}

	for l, p := range s.obj.Ports {
		if p.Port.Expose == port {
			s.obj.Ports = append(s.obj.Ports[:l], s.obj.Ports[l+1:]...)
			return nil
		}
	}

	return fmt.Errorf("port %d not configured on service", port)
}

// AddNodePort will set the node port on the service
func (s *Service) AddNodePort(port int32) {
	s.obj.NodePort = port
}

// RemoveNodePort will remove the node port from the service
func (s *Service) RemoveNodePort() {
	s.obj.NodePort = 0
}

// AddExternalIPs will add the provided external ip address to the service
func (s *Service) AddExternalIPs(ips []string) {
	for _, ip := range ips {
		s.obj.ExternalIPs = append(s.obj.ExternalIPs, types.IPAddr(ip))
	}
}

// RemoveExternalIP will remove the provided external ip address from the service
func (s *Service) RemoveExternalIP(ip string) {
	for l, sip := range s.obj.ExternalIPs {
		if strings.Compare(string(sip), ip) == 0 {
			s.obj.ExternalIPs = append(s.obj.ExternalIPs[:l], s.obj.ExternalIPs[l+1:]...)
			break
		}
	}
}

// AddLoadBalancer will add the provided load balancer to the service
func (s *Service) AddLoadBalancer(config api.LoadBalancerConfig) error {
	ips := []types.CIDR{}
	for _, ip := range config.AllowedIPs {
		ips = append(ips, types.CIDR(ip))
	}

	ingressList := []types.LoadBalancerIngress{}
	for _, ic := range config.Ingress {
		ingressLB, err := createIngressConfig(ic)
		if err != nil {
			return fmt.Errorf("unable to create ingress configuration: %v", err)
		}
		ingressList = append(ingressList, *ingressLB)
	}

	ingress := types.LoadBalancer{
		IP:                  types.IPAddr(config.IP),
		Allowed:             ips,
		HealthCheckNodePort: config.HealthCheckNodePort,
		Ingress:             ingressList,
	}

	s.obj.SetLoadBalancer(&ingress)
	return nil
}

// RemoveLoadBalancer will remove the load balancer from the service
func (s *Service) RemoveLoadBalancer() {
	ingress := types.LoadBalancer{
		IP:                  "",
		Allowed:             []types.CIDR{},
		HealthCheckNodePort: 0,
		Ingress:             []types.LoadBalancerIngress{},
	}

	s.obj.SetLoadBalancer(&ingress)
}

// ToKube returns the kubernetes version of the service
func (s *Service) ToKube() (runtime.Object, error) {
	wrapper := &types.ServiceWrapper{Service: *s.obj}
	return converters.Convert_Koki_Service_To_Kube_v1_Service(wrapper)
}

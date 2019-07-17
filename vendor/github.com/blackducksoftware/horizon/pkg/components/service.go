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
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/imdario/mergo"
)

// Service defines the service component
type Service struct {
	*v1.Service
	MetadataFuncs
}

// NewService creates a Service object
func NewService(config api.ServiceConfig) *Service {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	s := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.ServiceSpec{
			ClusterIP:                config.ClusterIP,
			ExternalName:             config.ExternalName,
			PublishNotReadyAddresses: config.PublishNotReadyAddresses,
		},
	}

	if config.IPTimeout != nil && *config.IPTimeout > 0 {
		s.Spec.SessionAffinityConfig = &v1.SessionAffinityConfig{
			ClientIP: &v1.ClientIPConfig{
				TimeoutSeconds: config.IPTimeout,
			},
		}
	}

	switch config.Affinity {
	case api.ServiceAffinityTypeClientIP:
		s.Spec.SessionAffinity = v1.ServiceAffinityClientIP
	case api.ServiceAffinityTypeNone:
		s.Spec.SessionAffinity = v1.ServiceAffinityNone
	}

	switch config.Type {
	case api.ServiceTypeServiceIP:
		s.Spec.Type = v1.ServiceTypeClusterIP
	case api.ServiceTypeNodePort:
		s.Spec.Type = v1.ServiceTypeNodePort
	case api.ServiceTypeLoadBalancer:
		s.Spec.Type = v1.ServiceTypeLoadBalancer
	case api.ServiceTypeExternalName:
		s.Spec.Type = v1.ServiceTypeExternalName
	}

	switch config.TrafficPolicy {
	case api.ServiceTrafficPolicyLocal:
		s.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
	case api.ServiceTrafficPolicyCluster:
		s.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
	}

	return &Service{&s, MetadataFuncs{&s}}
}

// AddSelectors adds selectors to the service
func (s *Service) AddSelectors(new map[string]string) {
	mergo.Merge(&s.Spec.Selector, new, mergo.WithOverride)
}

// RemoveSelectors removes selectors from the service
func (s *Service) RemoveSelectors(remove []string) {
	for _, k := range remove {
		delete(s.Spec.Selector, k)
	}
}

// AddPort adds a port to the service.  There can only be 1 non-named
// port on the service.  If there are more than 1 port then they must all
// be named.  If an unnamed port is already added to the service and another
// port is attmepted to be added, this function will return an error
func (s *Service) AddPort(config api.ServicePortConfig) error {
	if len(config.Name) == 0 && len(s.Spec.Ports) > 0 {
		return fmt.Errorf("all ports must be named if there is more than 1 port")
	}

	sp := v1.ServicePort{
		Name:       config.Name,
		Protocol:   convertProtocol(config.Protocol),
		Port:       config.Port,
		TargetPort: intstr.Parse(config.TargetPort),
		NodePort:   config.NodePort,
	}

	s.Spec.Ports = append(s.Spec.Ports, sp)

	return nil
}

// RemovePort will remove a port from the service
func (s *Service) RemovePort(port int32) error {
	for l, p := range s.Spec.Ports {
		if p.Port == port {
			s.Spec.Ports = append(s.Spec.Ports[:l], s.Spec.Ports[l+1:]...)
			return nil
		}
	}

	return fmt.Errorf("port %d not configured on service", port)
}

// AddExternalIPs will add the provided external ip address to the service
func (s *Service) AddExternalIPs(ips []string) {
	for _, ip := range ips {
		s.Spec.ExternalIPs = append(s.Spec.ExternalIPs, ip)
	}
}

// RemoveExternalIP will remove the provided external ip address from the service
func (s *Service) RemoveExternalIP(ip string) {
	for l, sip := range s.Spec.ExternalIPs {
		if strings.EqualFold(sip, ip) {
			s.Spec.ExternalIPs = append(s.Spec.ExternalIPs[:l], s.Spec.ExternalIPs[l+1:]...)
			break
		}
	}
}

// AddLoadBalancer will add the provided load balancer to the service
func (s *Service) AddLoadBalancer(config api.LoadBalancerConfig) error {
	if s.Spec.Type != v1.ServiceTypeLoadBalancer {
		return fmt.Errorf("unable to add a load balancer if service type isn't load balancer")
	}

	s.Spec.LoadBalancerIP = config.IP
	s.Spec.LoadBalancerSourceRanges = config.AllowedIPs
	s.Spec.HealthCheckNodePort = config.HealthCheckNodePort

	return nil
}

// RemoveLoadBalancer will remove the load balancer from the service
func (s *Service) RemoveLoadBalancer() {
	if s.Spec.Type == v1.ServiceTypeLoadBalancer {
		s.Spec.LoadBalancerIP = ""
		s.Spec.LoadBalancerSourceRanges = []string{}
		s.Spec.HealthCheckNodePort = 0
	}
}

// Deploy will deploy the service to the cluster
func (s *Service) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().Services(s.Namespace).Create(s.Service)
	return err
}

// Undeploy will remove the service from the cluster
func (s *Service) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().Services(s.Namespace).Delete(s.Name, &metav1.DeleteOptions{})
}

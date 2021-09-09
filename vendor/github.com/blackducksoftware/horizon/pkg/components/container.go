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
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Container defines containers that can be added to other components
type Container struct {
	*v1.Container
}

// NewContainer creates a Container object
func NewContainer(config api.ContainerConfig) (*Container, error) {
	limits := v1.ResourceList{}
	requests := v1.ResourceList{}

	if len(config.MinMem) > 0 {
		minMem, err := resource.ParseQuantity(config.MinMem)
		if err != nil {
			return nil, err
		}
		requests[v1.ResourceMemory] = minMem
	}

	if len(config.MaxMem) > 0 {
		maxMem, err := resource.ParseQuantity(config.MaxMem)
		if err != nil {
			return nil, err
		}
		limits[v1.ResourceMemory] = maxMem
	}

	if len(config.MinCPU) > 0 {
		minCPU, err := resource.ParseQuantity(config.MinCPU)
		if err != nil {
			return nil, err
		}
		requests[v1.ResourceCPU] = minCPU
	}

	if len(config.MaxCPU) > 0 {
		maxCPU, err := resource.ParseQuantity(config.MaxCPU)
		if err != nil {
			return nil, err
		}
		limits[v1.ResourceCPU] = maxCPU
	}

	c := v1.Container{
		Name:                   config.Name,
		Image:                  config.Image,
		Command:                config.Command,
		Args:                   config.Args,
		WorkingDir:             config.WorkingDirectory,
		TerminationMessagePath: config.TerminationMsgPath,
		Stdin:                  config.AllocateStdin,
		StdinOnce:              config.StdinOnce,
		TTY:                    config.AllocateTTY,
		Resources: v1.ResourceRequirements{
			Limits:   limits,
			Requests: requests,
		},
		SecurityContext: createSecurityContext(config),
	}

	switch config.TerminationMsgPolicy {
	case api.TerminationMessageReadFile:
		c.TerminationMessagePolicy = v1.TerminationMessageReadFile
	case api.TerminationMessageFallbackToLogsOnError:
		c.TerminationMessagePolicy = v1.TerminationMessageFallbackToLogsOnError
	}

	switch config.PullPolicy {
	case api.PullAlways:
		c.ImagePullPolicy = v1.PullAlways
	case api.PullNever:
		c.ImagePullPolicy = v1.PullNever
	case api.PullIfNotPresent:
		c.ImagePullPolicy = v1.PullIfNotPresent
	}

	return &Container{&c}, nil
}

func createSecurityContext(conf api.ContainerConfig) *v1.SecurityContext {
	sc := &v1.SecurityContext{}

	sc.Privileged = conf.Privileged
	sc.AllowPrivilegeEscalation = conf.AllowPrivilegeEscalation
	sc.ReadOnlyRootFilesystem = conf.ReadOnlyFS
	sc.RunAsNonRoot = conf.ForceNonRoot
	sc.RunAsUser = conf.UID
	sc.RunAsGroup = conf.GID
	sc.SELinuxOptions = createSELinux(conf.SELinux)

	if conf.ProcMount != nil {
		var pm v1.ProcMountType

		switch *conf.ProcMount {
		case api.ProcMountTypeDefault:
			pm = v1.DefaultProcMount
		case api.ProcMountTypeUmasked:
			pm = v1.UnmaskedProcMount
		}

		if len(string(pm)) > 0 {
			sc.ProcMount = &pm
		}
	}

	if reflect.DeepEqual(sc, &v1.SecurityContext{}) {
		return nil
	}

	return sc

}

// AddSELinux will add the provided SELinux configuration to the container
func (c *Container) AddSELinux(config api.SELinuxType) {
	if c.SecurityContext == nil {
		c.SecurityContext = &v1.SecurityContext{}
	}
	c.SecurityContext.SELinuxOptions = createSELinux(&config)
}

// RemoveSELinux will remove the SELinux configuration from the container
func (c *Container) RemoveSELinux() {
	if c.SecurityContext != nil {
		c.SecurityContext.SELinuxOptions = nil
	}
}

// AddVolumeMount adds a volume mount to the container
func (c *Container) AddVolumeMount(config api.VolumeMountConfig) error {
	if loc := c.findVolumeMount(config, c.VolumeMounts); loc >= 0 {
		return fmt.Errorf("volume mount %v already exists", config)
	}

	vm := v1.VolumeMount{
		MountPath: config.MountPath,
		Name:      config.Name,
		SubPath:   config.SubPath,
		ReadOnly:  config.ReadOnly,
	}

	if config.Propagation != nil {
		var propagation v1.MountPropagationMode
		switch *config.Propagation {
		case api.MountPropagationHostToContainer:
			propagation = v1.MountPropagationHostToContainer
		case api.MountPropagationBidirectional:
			propagation = v1.MountPropagationBidirectional
		case api.MountPropagationNone:
			propagation = v1.MountPropagationNone
		}

		vm.MountPropagation = &propagation
	}

	c.VolumeMounts = append(c.VolumeMounts, vm)
	return nil
}

// RemoveVolumeMount removes a volume mount from the container
func (c *Container) RemoveVolumeMount(config api.VolumeMountConfig) error {
	loc := c.findVolumeMount(config, c.VolumeMounts)
	if loc < 0 {
		return fmt.Errorf("volume mount %v doesn't exist", config)
	}

	c.VolumeMounts = append(c.VolumeMounts[:loc], c.VolumeMounts[loc+1:]...)
	return nil
}

func (c *Container) findVolumeMount(mountConfig api.VolumeMountConfig, mounts []v1.VolumeMount) int {
	for i, m := range mounts {
		if strings.EqualFold(mountConfig.Name, m.Name) &&
			strings.EqualFold(mountConfig.MountPath, m.MountPath) &&
			strings.EqualFold(mountConfig.SubPath, m.SubPath) {
			return i
		}
	}

	return -1
}

// AddVolumeDevice adds a volume device to the container
func (c *Container) AddVolumeDevice(config api.VolumeDeviceConfig) error {
	if loc := c.findVolumeDevice(config.Name, c.VolumeDevices); loc >= 0 {
		return fmt.Errorf("volume device %s already exists", config.Name)
	}

	d := v1.VolumeDevice{
		Name:       config.Name,
		DevicePath: config.Path,
	}

	c.VolumeDevices = append(c.VolumeDevices, d)
	return nil
}

// RemoveVolumeDevice removes a volume device from the container
func (c *Container) RemoveVolumeDevice(name string) error {
	loc := c.findVolumeDevice(name, c.VolumeDevices)
	if loc < 0 {
		return fmt.Errorf("volume device with name %s doesn't exist", name)
	}

	c.VolumeDevices = append(c.VolumeDevices[:loc], c.VolumeDevices[loc+1:]...)
	return nil
}

func (c *Container) findVolumeDevice(name string, devices []v1.VolumeDevice) int {
	for i, d := range devices {
		if strings.Compare(d.Name, name) == 0 {
			return i
		}
	}

	return -1
}

// AddEnv adds an environment configuration to the container
func (c *Container) AddEnv(config api.EnvConfig) {
	var env v1.EnvVar
	var envFrom v1.EnvFromSource

	switch config.Type {
	case api.EnvFromConfigMap:
		if len(config.KeyOrVal) > 0 {
			env = v1.EnvVar{
				Name: config.NameOrPrefix,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: config.FromName,
						},
						Key:      config.KeyOrVal,
						Optional: config.Optional,
					},
				},
			}
		} else {
			envFrom = v1.EnvFromSource{
				Prefix: config.NameOrPrefix,
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: config.FromName,
					},
					Optional: config.Optional,
				},
			}
		}

	case api.EnvFromSecret:
		if len(config.KeyOrVal) > 0 {
			env = v1.EnvVar{
				Name: config.NameOrPrefix,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: config.FromName,
						},
						Key:      config.KeyOrVal,
						Optional: config.Optional,
					},
				},
			}
		} else {
			envFrom = v1.EnvFromSource{
				Prefix: config.NameOrPrefix,
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: config.FromName,
					},
					Optional: config.Optional,
				},
			}
		}

	case api.EnvFromCPULimits:
		env = c.createResourceFieldRefEnv("limits.cpu", config.NameOrPrefix)

	case api.EnvFromMemLimits:
		env = c.createResourceFieldRefEnv("limits.memory", config.NameOrPrefix)

	case api.EnvFromEphemeralStorageLimits:
		env = c.createResourceFieldRefEnv("limits.ephemeral-storage", config.NameOrPrefix)

	case api.EnvFromCPURequests:
		env = c.createResourceFieldRefEnv("requests.cpu", config.NameOrPrefix)

	case api.EnvFromMemRequests:
		env = c.createResourceFieldRefEnv("requests.memory", config.NameOrPrefix)

	case api.EnvFromEphemeralStorageRequests:
		env = c.createResourceFieldRefEnv("requests.ephemeral-storage", config.NameOrPrefix)

	case api.EnvFromName:
		env = c.createFieldRefEnv("metadata.name", config.NameOrPrefix)

	case api.EnvFromNamespace:
		env = c.createFieldRefEnv("metadata.namespace", config.NameOrPrefix)

	case api.EnvFromLabels:
		env = c.createFieldRefEnv("metadata.labels", config.NameOrPrefix)

	case api.EnvFromAnnotation:
		env = c.createFieldRefEnv("metadata.annotations", config.NameOrPrefix)

	case api.EnvFromNodename:
		env = c.createFieldRefEnv("spec.nodeName", config.NameOrPrefix)

	case api.EnvFromServiceAccountName:
		env = c.createFieldRefEnv("spec.serviceAccountName", config.NameOrPrefix)

	case api.EnvFromHostIP:
		env = c.createFieldRefEnv("status.hostIP", config.NameOrPrefix)

	case api.EnvFromPodIP:
		env = c.createFieldRefEnv("status.podIP", config.NameOrPrefix)

	case api.EnvVal:
		env = v1.EnvVar{
			Name:  config.NameOrPrefix,
			Value: config.KeyOrVal,
		}
	}

	if !reflect.DeepEqual(env, v1.EnvVar{}) {
		c.Env = append(c.Env, env)
	}

	if !reflect.DeepEqual(envFrom, v1.EnvFromSource{}) {
		c.EnvFrom = append(c.EnvFrom, envFrom)
	}
}

func (c *Container) createResourceFieldRefEnv(refType string, name string) v1.EnvVar {
	// ResourceFieldRef
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			ResourceFieldRef: &v1.ResourceFieldSelector{
				Resource: refType,
			},
		},
	}
}

func (c *Container) createFieldRefEnv(refType string, name string) v1.EnvVar {
	// FieldRef
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: refType,
			},
		},
	}
}

// AddPort adds a port to expose to the container
func (c *Container) AddPort(config api.PortConfig) {
	p := v1.ContainerPort{
		Name:          config.Name,
		HostIP:        config.IP,
		HostPort:      config.HostPort,
		ContainerPort: config.ContainerPort,
		Protocol:      convertProtocol(config.Protocol),
	}

	c.Ports = append(c.Ports, p)
}

// RemovePort removes an exposed port from the container
func (c *Container) RemovePort(name string) error {
	loc := c.findPort(name, c.Ports)
	if loc < 0 {
		return fmt.Errorf("port with name %s doesn't exist", name)
	}

	c.Ports = append(c.Ports[:loc], c.Ports[loc+1:]...)
	return nil

}

func (c *Container) findPort(name string, ports []v1.ContainerPort) int {
	for i, p := range ports {
		if strings.Compare(p.Name, name) == 0 {
			return i
		}
	}

	return -1
}

// AddLivenessProbe adds a liveness probe to the container
func (c *Container) AddLivenessProbe(config api.ProbeConfig) {
	c.LivenessProbe = c.createProbe(config)
}

// RemoveLivenessProbe remove the existing liveness probe from the
// container if it exists
func (c *Container) RemoveLivenessProbe() {
	c.LivenessProbe = nil
}

// AddReadinessProbe adds a readiness probe to the container
func (c *Container) AddReadinessProbe(config api.ProbeConfig) {
	c.ReadinessProbe = c.createProbe(config)
}

// RemoveReadinessProbe remove the existing readiness probe from the
// container if it exists
func (c *Container) RemoveReadinessProbe() {
	c.ReadinessProbe = nil
}

func (c *Container) createProbe(config api.ProbeConfig) *v1.Probe {
	h := c.createHandler(config.ActionConfig)
	p := &v1.Probe{
		Handler:             *h,
		InitialDelaySeconds: config.Delay,
		PeriodSeconds:       config.Interval,
		SuccessThreshold:    config.MinCountSuccess,
		FailureThreshold:    config.MinCountFailure,
		TimeoutSeconds:      config.Timeout,
	}

	return p
}

func (c *Container) createHandler(config api.ActionConfig) *v1.Handler {
	handler := &v1.Handler{}
	switch config.Type {
	case api.ActionTypeCommand:
		handler.Exec = &v1.ExecAction{
			Command: config.Command,
		}

	case api.ActionTypeHTTP, api.ActionTypeHTTPS:
		var scheme v1.URIScheme

		if config.Type == api.ActionTypeHTTP {
			scheme = v1.URISchemeHTTP
		} else {
			scheme = v1.URISchemeHTTPS
		}

		port := intstr.FromString("80")
		if len(config.Port) > 0 {
			port = intstr.Parse(config.Port)
		}

		var headers []v1.HTTPHeader
		for k, v := range config.Headers {
			kubeHeader := v1.HTTPHeader{
				Name:  k,
				Value: v,
			}
			headers = append(headers, kubeHeader)
		}

		handler.HTTPGet = &v1.HTTPGetAction{
			Scheme:      scheme,
			Path:        config.Path,
			Port:        port,
			Host:        config.Host,
			HTTPHeaders: headers,
		}

	case api.ActionTypeTCP:
		port := intstr.FromString("80")
		if len(config.Port) > 0 {
			port = intstr.Parse(config.Port)
		}

		handler.TCPSocket = &v1.TCPSocketAction{
			Host: config.Host,
			Port: port,
		}
	}

	return handler
}

// AddPostStartAction adds a PostStart action to the container
func (c *Container) AddPostStartAction(config api.ActionConfig) {
	if c.Lifecycle == nil {
		c.Lifecycle = &v1.Lifecycle{}
	}
	c.Lifecycle.PostStart = c.createHandler(config)
}

// RemovePostStartAction removes the PostStart action from the container
// if it exists
func (c *Container) RemovePostStartAction() {
	if c.Lifecycle != nil {
		c.Lifecycle.PostStart = nil
		if c.Lifecycle.PreStop == nil {
			c.Lifecycle = nil
		}
	}
}

// AddPreStopAction adds a PreStop action to the container
func (c *Container) AddPreStopAction(config api.ActionConfig) {
	if c.Lifecycle == nil {
		c.Lifecycle = &v1.Lifecycle{}
	}
	c.Lifecycle.PreStop = c.createHandler(config)
}

// RemovePreStopAction removes the PreStop action from the container
// if it exists
func (c *Container) RemovePreStopAction() {
	if c.Lifecycle != nil {
		c.Lifecycle.PreStop = nil
		if c.Lifecycle.PostStart == nil {
			c.Lifecycle = nil
		}
	}
}

// AddAddCapabilities will add POSIX capabilities that will be
// added to the container
func (c *Container) AddAddCapabilities(add []string) {
	if c.SecurityContext == nil {
		c.SecurityContext = &v1.SecurityContext{}
	}

	if c.SecurityContext.Capabilities == nil {
		c.SecurityContext.Capabilities = &v1.Capabilities{}
	}

	for _, cap := range add {
		c.SecurityContext.Capabilities.Add = c.appendCapabilityIfMissing(v1.Capability(cap), c.SecurityContext.Capabilities.Add)
	}
}

// RemoveAddCapability will remove a POSIX capabiltiy that would be
// added to the container
func (c *Container) RemoveAddCapability(remove string) {
	if c.SecurityContext != nil && c.SecurityContext.Capabilities != nil {
		for l, cap := range c.SecurityContext.Capabilities.Add {
			if strings.EqualFold(string(cap), remove) {
				c.SecurityContext.Capabilities.Add = append(c.SecurityContext.Capabilities.Add[:l], c.SecurityContext.Capabilities.Add[l+1:]...)
				break
			}
		}
	}
}

// AddDeleteCapabilities will add POSIX capabilities that will be
// removed from the container
func (c *Container) AddDeleteCapabilities(add []string) {
	if c.SecurityContext == nil {
		c.SecurityContext = &v1.SecurityContext{}
	}

	if c.SecurityContext.Capabilities == nil {
		c.SecurityContext.Capabilities = &v1.Capabilities{}
	}

	for _, cap := range add {
		c.SecurityContext.Capabilities.Drop = c.appendCapabilityIfMissing(v1.Capability(cap), c.SecurityContext.Capabilities.Drop)
	}
}

// RemoveDeleteCapability will remove a POSIX capabiltiy will be
// removed from the container
func (c *Container) RemoveDeleteCapability(remove string) {
	if c.SecurityContext != nil && c.SecurityContext.Capabilities != nil {
		for l, cap := range c.SecurityContext.Capabilities.Drop {
			if strings.EqualFold(string(cap), remove) {
				c.SecurityContext.Capabilities.Drop = append(c.SecurityContext.Capabilities.Drop[:l], c.SecurityContext.Capabilities.Drop[l+1:]...)
				break
			}
		}
	}
}

func (c *Container) appendCapabilityIfMissing(new v1.Capability, list []v1.Capability) []v1.Capability {
	for _, cap := range list {
		if strings.EqualFold(string(new), string(cap)) {
			return list
		}
	}

	return append(list, new)
}

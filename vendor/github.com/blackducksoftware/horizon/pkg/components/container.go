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

	"github.com/koki/short/types"
	"github.com/koki/short/util/floatstr"
)

// Container defines containers that can be added to other components
type Container struct {
	obj *types.Container
}

// NewContainer creates a Container object
func NewContainer(config api.ContainerConfig) *Container {
	args := []floatstr.FloatOrString{}
	for _, a := range config.Args {
		newArg := floatstr.Parse(a)
		args = append(args, *newArg)
	}

	c := &types.Container{
		Name:               config.Name,
		Command:            config.Command,
		Args:               args,
		Image:              config.Image,
		Privileged:         config.Privileged,
		AllowEscalation:    config.AllowPrivilegeEscalation,
		ForceNonRoot:       config.ForceNonRoot,
		UID:                config.UID,
		Stdin:              config.AllocateStdin,
		StdinOnce:          config.StdinOnce,
		TTY:                config.AllocateTTY,
		WorkingDir:         config.WorkingDirectory,
		TerminationMsgPath: config.TerminationMsgPath,
	}

	switch config.TerminationMsgPolicy {
	case api.TerminationMessageReadFile:
		c.TerminationMsgPolicy = types.TerminationMessageReadFile
	case api.TerminationMessageFallbackToLogsOnError:
		c.TerminationMsgPolicy = types.TerminationMessageFallbackToLogsOnError
	}

	if config.ReadOnlyFS != nil {
		if *config.ReadOnlyFS == true {
			ro := true
			c.RO = &ro
		} else {
			rw := true
			c.RW = &rw
		}
	}

	if len(config.MinCPU) > 0 || len(config.MaxCPU) > 0 {
		cpu := types.CPU{
			Min: config.MinCPU,
			Max: config.MaxCPU,
		}
		c.CPU = &cpu
	}

	if len(config.MinMem) > 0 || len(config.MaxMem) > 0 {
		mem := types.Mem{
			Min: config.MinMem,
			Max: config.MaxMem,
		}
		c.Mem = &mem
	}

	switch config.PullPolicy {
	case api.PullAlways:
		c.Pull = types.PullAlways
	case api.PullNever:
		c.Pull = types.PullNever
	case api.PullIfNotPresent:
		c.Pull = types.PullIfNotPresent
	}

	container := Container{obj: c}

	if config.SELinux != nil {
		container.AddSELinux(*config.SELinux)
	}
	return &container
}

// AddSELinux will add the provided SELinux configuration to the container
func (c *Container) AddSELinux(config api.SELinuxType) {
	c.obj.SELinux = createSELinuxObj(config)
}

// RemoveSELinux will remove the SELinux configuration from the container
func (c *Container) RemoveSELinux() {
	c.obj.SELinux = nil
}

// AddVolumeMount adds a volume mount to the container
func (c *Container) AddVolumeMount(config api.VolumeMountConfig) error {
	if loc := c.findVolumeMount(config.Name, c.obj.VolumeMounts); loc >= 0 {
		return fmt.Errorf("volume mount %s already exists", config.Name)
	}

	store := config.Name
	if len(config.SubPath) != 0 {
		store += fmt.Sprintf(":%s", config.SubPath)
	}

	if config.ReadOnly != nil && *config.ReadOnly == true {
		store += ":ro"
	}

	vm := types.VolumeMount{
		MountPath: config.MountPath,
		Store:     store,
	}

	if config.Propagation != nil {
		var propagation types.MountPropagation
		switch *config.Propagation {
		case api.MountPropagationHostToContainer:
			propagation = types.MountPropagationHostToContainer
			vm.Propagation = &propagation
		case api.MountPropagationBidirectional:
			propagation = types.MountPropagationBidirectional
			vm.Propagation = &propagation
		case api.MountPropagationNone:
			propagation = types.MountPropagationNone
			vm.Propagation = &propagation
		}
	}

	c.obj.VolumeMounts = append(c.obj.VolumeMounts, vm)
	return nil
}

// RemoveVolumeMount removes a volume mount from the container
func (c *Container) RemoveVolumeMount(name string) error {
	loc := c.findVolumeMount(name, c.obj.VolumeMounts)
	if loc < 0 {
		return fmt.Errorf("volume mount with name %s doesn't exist", name)
	}

	c.obj.VolumeMounts = append(c.obj.VolumeMounts[:loc], c.obj.VolumeMounts[loc+1:]...)
	return nil
}

func (c *Container) findVolumeMount(name string, mounts []types.VolumeMount) int {
	for i, c := range mounts {
		if strings.Compare(c.Store, name) == 0 {
			return i
		}
	}

	return -1
}

// GetObj returns the container object in a format the deployer
// can use
func (c *Container) GetObj() *types.Container {
	return c.obj
}

// GetName retruns the name of the container
func (c *Container) GetName() string {
	return c.obj.Name
}

// AddEnv adds an environment configuration to the container
func (c *Container) AddEnv(config api.EnvConfig) error {
	var env types.Env
	var err error

	switch config.Type {
	case api.EnvFromConfigMap:
		env, err = types.NewEnvFromConfig(config.NameOrPrefix, config.FromName, config.KeyOrVal)
	case api.EnvFromSecret:
		env, err = types.NewEnvFromSecret(config.NameOrPrefix, config.FromName, config.KeyOrVal)
	case api.EnvFromCPULimits:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeCPULimits)
	case api.EnvFromMemLimits:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMemLimits)
	case api.EnvFromEphemeralStorageLimits:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeEphemeralStorageLimits)
	case api.EnvFromCPURequests:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeCPURequests)
	case api.EnvFromMemRequests:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMemRequests)
	case api.EnvFromEphemeralStorageRequests:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeEphemeralStorageRequests)
	case api.EnvFromName:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMetadataName)
	case api.EnvFromNamespace:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMetadataNamespace)
	case api.EnvFromLabels:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMetadataLabels)
	case api.EnvFromAnnotation:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeMetadataAnnotation)
	case api.EnvFromNodename:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeSpecNodename)
	case api.EnvFromServiceAccountName:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeSpecServiceAccountName)
	case api.EnvFromHostIP:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeStatusHostIP)
	case api.EnvFromPodIP:
		env, err = types.NewEnvFrom(config.NameOrPrefix, types.EnvFromTypeStatusPodIP)
	case api.EnvVal:
		env, err = types.NewEnv(config.NameOrPrefix, config.KeyOrVal)
	}

	if err != nil {
		return err
	}

	c.obj.Env = append(c.obj.Env, env)
	return nil
}

// AddPort adds a port to expose to the container
func (c *Container) AddPort(config api.PortConfig) {
	p := types.Port{
		Name:          config.Name,
		IP:            config.IP,
		HostPort:      config.HostPort,
		ContainerPort: config.ContainerPort,
	}

	switch config.Protocol {
	case api.ProtocolTCP:
		p.Protocol = types.ProtocolTCP
	case api.ProtocolUDP:
		p.Protocol = types.ProtocolUDP
	}

	c.obj.Expose = append(c.obj.Expose, p)
}

// RemovePort removes an exposed port from the container
func (c *Container) RemovePort(name string) error {
	loc := c.findPort(name, c.obj.Expose)
	if loc < 0 {
		return fmt.Errorf("port with name %s doesn't exist", name)
	}

	c.obj.Expose = append(c.obj.Expose[:loc], c.obj.Expose[loc+1:]...)
	return nil

}

func (c *Container) findPort(name string, ports []types.Port) int {
	for i, p := range ports {
		if strings.Compare(p.Name, name) == 0 {
			return i
		}
	}

	return -1
}

// AddLivenessProbe adds a liveness probe to the container
func (c *Container) AddLivenessProbe(config api.ProbeConfig) {
	c.obj.LivenessProbe = c.createProbe(config)
}

// RemoveLivenessProbe remove the existing liveness probe from the
// container if it exists
func (c *Container) RemoveLivenessProbe() {
	if c.obj.LivenessProbe != nil {
		c.obj.LivenessProbe = nil
	}
}

// AddReadinessProbe adds a readiness probe to the container
func (c *Container) AddReadinessProbe(config api.ProbeConfig) {
	c.obj.ReadinessProbe = c.createProbe(config)
}

// RemoveReadinessProbe remove the existing readiness probe from the
// container if it exists
func (c *Container) RemoveReadinessProbe() {
	if c.obj.ReadinessProbe != nil {
		c.obj.ReadinessProbe = nil
	}
}

func (c *Container) createProbe(config api.ProbeConfig) *types.Probe {
	a := c.createAction(config.ActionConfig)
	p := &types.Probe{
		Action:          *a,
		Delay:           config.Delay,
		Interval:        config.Interval,
		MinCountSuccess: config.MinCountSuccess,
		MinCountFailure: config.MinCountFailure,
		Timeout:         config.Timeout,
	}

	return p
}

func (c *Container) createAction(config api.ActionConfig) *types.Action {
	a := types.Action{
		Command: config.Command,
	}
	if len(config.URL) > 0 {
		var headers []string
		for _, h := range config.Headers {
			headers = append(headers, fmt.Sprintf("%s:%s", h.Name, h.Value))
		}
		na := types.NetAction{
			Headers: headers,
			URL:     config.URL,
		}
		a.Net = &na
	}
	return &a
}

// AddPostStartAction adds a PostStart action to the container
func (c *Container) AddPostStartAction(config api.ActionConfig) {
	c.obj.OnStart = c.createAction(config)
}

// RemovePostStartAction removes the PostStart action from the container
// if it exists
func (c *Container) RemovePostStartAction() {
	if c.obj.OnStart != nil {
		c.obj.OnStart = nil
	}
}

// AddPreStopAction adds a PreStop action to the container
func (c *Container) AddPreStopAction(config api.ActionConfig) {
	c.obj.PreStop = c.createAction(config)
}

// RemovePreStopAction removes the PreStop action from the container
// if it exists
func (c *Container) RemovePreStopAction() {
	if c.obj.PreStop != nil {
		c.obj.PreStop = nil
	}
}

// AddAddCapabilities will add POSIX capabilities that will be
// added to the container
func (c *Container) AddAddCapabilities(add []string) {
	for _, cap := range add {
		c.obj.AddCapabilities = appendIfMissing(cap, c.obj.AddCapabilities)
	}
}

// RemoveAddCapability will remove a POSIX capabiltiy that would be
// added to the container
func (c *Container) RemoveAddCapability(remove string) {
	for l, cap := range c.obj.AddCapabilities {
		if strings.Compare(cap, remove) == 0 {
			c.obj.AddCapabilities = append(c.obj.AddCapabilities[:l], c.obj.AddCapabilities[l+1:]...)
			break
		}
	}
}

// AddDeleteCapabilities will add POSIX capabilities that will be
// removed from the container
func (c *Container) AddDeleteCapabilities(add []string) {
	for _, cap := range add {
		c.obj.DelCapabilities = appendIfMissing(cap, c.obj.DelCapabilities)
	}
}

// RemoveDeleteCapability will remove a POSIX capabiltiy will be
// removed from the container
func (c *Container) RemoveDeleteCapability(remove string) {
	for l, cap := range c.obj.DelCapabilities {
		if strings.Compare(cap, remove) == 0 {
			c.obj.DelCapabilities = append(c.obj.DelCapabilities[:l], c.obj.DelCapabilities[l+1:]...)
			break
		}
	}
}

func (c *Container) appendIfMissing(new string, list []string) []string {
	for _, o := range list {
		if strings.Compare(new, o) == 0 {
			return list
		}
	}
	return append(list, new)
}

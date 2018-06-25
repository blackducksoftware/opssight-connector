/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershii. The ASF licenses this file
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

package protoform

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"github.com/spf13/viper"

	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/koki/short/util/floatstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Installer handles deploying configured components to a cluster
type Installer struct {
	config                 protoformConfig
	replicationControllers []*types.ReplicationController
	pods                   []*types.Pod
	configMaps             []*types.ConfigMap
	services               []*types.Service

	client *kubernetes.Clientset
}

// NewInstaller creates a Installer object
func NewInstaller(defaults *api.ProtoformDefaults, path string) *Installer {
	i := Installer{}
	i.readConfig(path)
	i.setDefaults(defaults)
	i.prettyPrint(i.config)
	return &i
}

func (i *Installer) setDefaults(defaults *api.ProtoformDefaults) {
	configFields := reflect.ValueOf(&i.config).Elem()
	defaultFields := reflect.ValueOf(defaults).Elem()
	for cnt := 0; cnt < configFields.NumField(); cnt++ {
		fieldName := configFields.Type().Field(cnt).Name
		field := configFields.Field(cnt)
		defaultValue := defaultFields.FieldByName(fieldName)
		if defaultValue.IsValid() {
			switch configFields.Type().Field(cnt).Type.Kind().String() {
			case "string":
				if field.Len() == 0 {
					field.Set(defaultValue)
				}
			case "slice":
				if field.Len() == 0 {
					field.Set(defaultValue)
				}
			case "int":
				if field.Int() == 0 {
					field.Set(defaultValue)
				}
			}
		}
	}
}

// We don't dynamically reload.
// If users want to dynamically reload,
// they can update the individual perceptor containers configmaps.
func (i *Installer) readConfig(configPath string) {
	log.Print("*************** [protoform] initializing  ****************")
	viper.SetConfigFile(configPath)

	// these need to be set before we read in the config!
	viper.SetEnvPrefix("PCP")
	viper.BindEnv("HubUserPassword")
	if viper.GetString("hubuserpassword") == "" {
		viper.Debug()
		panic("No hub database password secret supplied.  Please inject PCP_HUBUSERPASSWORD as a secret and restart")
	}

	i.config.HubUserPasswordEnvVar = "PCP_HUBUSERPASSWORD"
	i.config.ViperSecret = "viper-secret"
	log.Print(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Print(" ^^ Didnt see a config file! Using defaults for everything")
	}

	internalRegistry := viper.GetStringSlice("InternalDockerRegistries")
	viper.Set("InternalDockerRegistries", internalRegistry)

	viper.Unmarshal(&i.config)
	log.Print("*************** [protoform] done reading in config ****************")

}

// addRC will add a replication controller to the list of replication controllers
// to be deployed
func (i *Installer) addRC(config api.ReplicationControllerConfig) {
	newRc := &types.ReplicationController{
		Name:     config.Name,
		Replicas: &config.Replicas,
		Selector: config.Selector,
		TemplateMetadata: &types.PodTemplateMeta{
			Labels: config.Labels,
		},
		PodTemplate: types.PodTemplate{
			Volumes:    config.Vols,
			Containers: config.Containers,
			Account:    config.ServiceAccount,
		},
	}
	i.replicationControllers = append(i.replicationControllers, newRc)
}

// AddPod will add a pod to the list of pods to be deployed
func (i *Installer) addPod(config api.PodConfig) {
	newPod := &types.Pod{
		PodTemplateMeta: types.PodTemplateMeta{
			Name:   config.Name,
			Labels: config.Labels,
		},
		PodTemplate: types.PodTemplate{
			Volumes:    config.Vols,
			Containers: config.Containers,
			Account:    config.ServiceAccount,
		},
	}
	i.pods = append(i.pods, newPod)
}

// addService will add a service to the list of services
// to be deployed
func (i *Installer) addService(config api.ServiceConfig) {
	ports := []types.NamedServicePort{}
	for k, v := range config.Ports {
		newPort := types.NamedServicePort{
			Name: k,
			Port: types.ServicePort{
				Expose: v,
			},
		}
		ports = append(ports, newPort)
	}

	newSvc := &types.Service{
		Name:     config.Name,
		Ports:    ports,
		Selector: config.Selector,
	}

	i.services = append(i.services, newSvc)
}

// AddConfigMap will add a config map to the list of
// config maps to be deployed
func (i *Installer) AddConfigMap(conf api.ConfigMapConfig) {
	configMap := &types.ConfigMap{
		Name:      conf.Name,
		Namespace: conf.Namespace,
		Data:      map[string]string{},
	}
	for k, v := range conf.Data {
		configMap.Data[k] = v
	}
	i.configMaps = append(i.configMaps, configMap)
}

// Run will start the installer
func (i *Installer) Run() {
	if !i.config.DryRun {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		} else {
			// creates the client
			i.client, err = kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
		}
	}

	i.init()
	i.deploy()
	log.Println("Entering pod listing loop!")

	// continually print out pod statuses .  can exit any time.  maybe use this as a stub for self testing.
	if !i.config.DryRun {
		for cnt := 0; cnt < 10; cnt++ {
			pods, _ := i.client.Core().Pods(i.config.Namespace).List(v1meta.ListOptions{})
			for _, pod := range pods.Items {
				log.Printf("Pod = %v -> %v", pod.Name, pod.Status.Phase)
			}
			log.Printf("***************")
			time.Sleep(10 * time.Second)
		}
	}

	return
}

func (i *Installer) init() {
	i.addDefaultServiceAccounts()
	isValid := i.sanityCheckServices()
	if isValid == false {
		panic("Please set the service accounts correctly!")
	}

	i.substituteDefaultImageVersion()
	i.createConfigMaps()
	i.addPerceptorResources()
}

func (i *Installer) substituteDefaultImageVersion() {
	if len(i.config.PerceptorContainerVersion) == 0 {
		i.config.PerceptorContainerVersion = i.config.DefaultVersion
	}
	if len(i.config.ScannerContainerVersion) == 0 {
		i.config.ScannerContainerVersion = i.config.DefaultVersion
	}
	if len(i.config.PerceiverContainerVersion) == 0 {
		i.config.PerceiverContainerVersion = i.config.DefaultVersion
	}
	if len(i.config.ImageFacadeContainerVersion) == 0 {
		i.config.ImageFacadeContainerVersion = i.config.DefaultVersion
	}
	if len(i.config.SkyfireContainerVersion) == 0 {
		i.config.SkyfireContainerVersion = i.config.DefaultVersion
	}
}

func (i *Installer) addDefaultServiceAccounts() {
	// TODO Viperize these env vars.
	if len(i.config.ServiceAccounts) == 0 {
		log.Println("NO SERVICE ACCOUNTS FOUND.  USING DEFAULTS: MAKE SURE THESE EXIST!")

		svcAccounts := map[string]string{
			// WARNING: These service accounts need to exist !
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner",
		}
		// TODO programatically validate rather then sanity check.
		i.prettyPrint(svcAccounts)
		i.config.ServiceAccounts = svcAccounts
	}
}

// This function creates the volumes and containers that need to be used for adding RC or Pod
func (i *Installer) createRcOrPod(descriptions []*ReplicationController) (map[string]types.Volume, []types.Container) {

	TheVolumes := map[string]types.Volume{}
	TheContainers := []types.Container{}
	addedMounts := map[string]string{}

	for _, desc := range descriptions {
		mounts := []types.VolumeMount{}

		for cfgMapName, cfgMapMount := range desc.ConfigMapMounts {
			log.Print("Adding config mounts now.")
			if addedMounts[cfgMapName] == "" {
				addedMounts[cfgMapName] = cfgMapName
				TheVolumes[cfgMapName] = types.Volume{
					ConfigMap: &types.ConfigMapVolume{
						Name: cfgMapName,
					},
				}
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", cfgMapName))
			}
			mounts = append(mounts, types.VolumeMount{
				Store:     cfgMapName,
				MountPath: cfgMapMount,
			})

		}

		// keep track of emptyDirs, only once, since it can be referenced in
		// multiple pods
		for emptyDirName, emptyDirMount := range desc.EmptyDirMounts {
			log.Print("Adding empty mounts now.")
			if addedMounts[emptyDirName] == "" {
				addedMounts[emptyDirName] = emptyDirName
				TheVolumes[emptyDirName] = types.Volume{
					EmptyDir: &types.EmptyDirVolume{
						Medium: types.StorageMediumDefault,
					},
				}
			} else {
				log.Print(fmt.Sprintf("Not adding volume, already added: %v", emptyDirName))
			}
			mounts = append(mounts, types.VolumeMount{
				Store:     emptyDirName,
				MountPath: emptyDirMount,
			})

		}

		if desc.DockerSocket {
			dockerSock := types.VolumeMount{
				Store:     "dir-docker-socket",
				MountPath: "/var/run/docker.sock",
			}
			mounts = append(mounts, dockerSock)
			TheVolumes[dockerSock.Store] = types.Volume{
				HostPath: &types.HostPathVolume{
					Path: dockerSock.MountPath,
				},
			}
		}

		for name := range desc.EmptyDirVolumeMounts {
			TheVolumes[name] = types.Volume{
				EmptyDir: &types.EmptyDirVolume{
					Medium: types.StorageMediumDefault,
				},
			}
		}

		envVar := []types.Env{}
		for _, env := range desc.Env {
			new, err := types.NewEnvFromSecret(env.EnvName, env.SecretName, env.KeyFromSecret)
			if err != nil {
				panic(err)
			}
			envVar = append(envVar, new)
		}

		container := types.Container{
			Name:    desc.Name,
			Image:   desc.Image,
			Pull:    types.PullAlways,
			Command: desc.Cmd,
			Args:    desc.Arg,
			Env:     envVar,
			Expose: []types.Port{
				{
					ContainerPort: fmt.Sprintf("%d", desc.Port),
					Protocol:      types.ProtocolTCP,
				},
			},
			CPU: &types.CPU{
				Min: desc.CPU.String(),
			},
			Mem: &types.Mem{
				Min: desc.Memory.String(),
			},
			VolumeMounts: mounts,
			Privileged:   &desc.DockerSocket,
		}
		// Each RC has only one pod, but can have many containers.
		TheContainers = append(TheContainers, container)

		log.Print(fmt.Sprintf("privileged = %v %v %v", desc.Name, desc.DockerSocket, *container.Privileged))
	}
	return TheVolumes, TheContainers
}

// AddReplicationControllerAndService function adds RC and services that forward to it to installation set
func (i *Installer) AddReplicationControllerAndService(descriptions []*ReplicationController) {
	// Add replicationController
	i.AddReplicationController(descriptions)
	// Add service
	i.AddService(descriptions)
}

// AddReplicationController function add a RC that forward to it to installation set
func (i *Installer) AddReplicationController(descriptions []*ReplicationController) {
	// Get the volumes and containers that will be used to create the replication contoller
	volumes, containers := i.createRcOrPod(descriptions)
	// Create the replication controller
	i.createRC(descriptions, volumes, containers)
}

// AddPod function add a Pod that forward to it to installation set
func (i *Installer) AddPod(descriptions []*ReplicationController) {
	// Get the volumes and containers that will be used to create the pod
	volumes, containers := i.createRcOrPod(descriptions)
	// Create the pod
	i.createPod(descriptions, volumes, containers)
}

// AddService function add a service that forward to it to installation set
func (i *Installer) AddService(descriptions []*ReplicationController) {
	// Create the service
	i.createService(descriptions)
}

// This function creates the pod config and adds to the pod list that forward to it to installation set
func (i *Installer) createPod(descriptions []*ReplicationController, volumes map[string]types.Volume, containers []types.Container) {
	podCfg := api.PodConfig{
		Name:           descriptions[0].Name,
		Labels:         map[string]string{"name": descriptions[0].Name},
		Vols:           volumes,
		Containers:     containers,
		ServiceAccount: descriptions[0].ServiceAccountName,
	}
	i.addPod(podCfg)
}

// This function creates the replication controller config and adds to the pod list that forward to it to installation set
func (i *Installer) createRC(descriptions []*ReplicationController, volumes map[string]types.Volume, containers []types.Container) {
	rcCfg := api.ReplicationControllerConfig{
		Name:           descriptions[0].Name,
		Replicas:       descriptions[0].Replicas,
		Selector:       map[string]string{"name": descriptions[0].Name},
		Labels:         map[string]string{"name": descriptions[0].Name},
		Vols:           volumes,
		Containers:     containers,
		ServiceAccount: descriptions[0].ServiceAccountName,
	}
	i.addRC(rcCfg)
}

// This function creates the service config and adds to the pod list that forward to it to installation set
func (i *Installer) createService(descriptions []*ReplicationController) {
	for _, desc := range descriptions {
		serviceCfg := api.ServiceConfig{
			Name:     desc.Name,
			Ports:    map[string]int32{desc.Name: desc.Port},
			Selector: map[string]string{"name": descriptions[0].Name},
		}
		i.addService(serviceCfg)
	}
}

// GenerateDefaultCPU function generates the default CPU
func (i *Installer) GenerateDefaultCPU(defaultCPU string) (resource.Quantity, error) {
	return resource.ParseQuantity(defaultCPU)
}

// GenerateDefaultMemory function generates the default memory
func (i *Installer) GenerateDefaultMemory(defaultMem string) (resource.Quantity, error) {
	return resource.ParseQuantity(defaultMem)
}

// GenerateArg function generates the argument that to be called
func (i *Installer) GenerateArg(stringVal string, floatval float64) floatstr.FloatOrString {
	if len(stringVal) > 0 {
		return floatstr.FloatOrString{Type: floatstr.String, StringVal: stringVal}
	}
	return floatstr.FloatOrString{Type: floatstr.Float, FloatVal: floatval}
}

func (i *Installer) addPerceptorResources() {
	paths := i.generateContainerPaths()

	// WARNING: THE SERVICE ACCOUNT IN THE FIRST CONTAINER IS USED FOR THE GLOBAL SVC ACCOUNT FOR ALL PODS !!!!!!!!!!!!!
	// MAKE SURE IF YOU NEED A SVC ACCOUNT THAT ITS IN THE FIRST CONTAINER...
	defaultMem, err := i.GenerateDefaultMemory(i.config.DefaultMem)
	if err != nil {
		panic(err)
	}
	defaultCPU, err := i.GenerateDefaultCPU(i.config.DefaultCPU)
	if err != nil {
		panic(err)
	}

	i.AddReplicationControllerAndService([]*ReplicationController{
		{
			Replicas:        1,
			ConfigMapMounts: map[string]string{"perceptor": "/etc/perceptor"},
			Env: []envSecret{
				{
					EnvName:       i.config.HubUserPasswordEnvVar,
					SecretName:    i.config.ViperSecret,
					KeyFromSecret: "HubUserPassword",
				},
			},
			Name:   i.config.PerceptorImageName,
			Image:  paths["perceptor"],
			Port:   int32(i.config.PerceptorPort),
			Cmd:    []string{"./perceptor"},
			Arg:    []floatstr.FloatOrString{i.GenerateArg("/etc/perceptor/perceptor.yaml", 0)},
			CPU:    defaultCPU,
			Memory: defaultMem,
		},
	})

	i.AddReplicationControllerAndService([]*ReplicationController{
		{
			Replicas:        1,
			ConfigMapMounts: map[string]string{"perceiver": "/etc/perceiver"},
			EmptyDirMounts: map[string]string{
				"logs": "/tmp",
			},
			Name:               i.config.PodPerceiverImageName,
			Image:              paths["pod-perceiver"],
			Port:               int32(i.config.PerceiverPort),
			Cmd:                []string{"./pod-perceiver"},
			Arg:                []floatstr.FloatOrString{i.GenerateArg("/etc/perceiver/perceiver.yaml", 0)},
			ServiceAccountName: i.config.ServiceAccounts["pod-perceiver"],
			ServiceAccount:     i.config.ServiceAccounts["pod-perceiver"],
			CPU:                defaultCPU,
			Memory:             defaultMem,
		},
	})

	i.AddReplicationControllerAndService([]*ReplicationController{
		{
			Replicas:        int32(math.Ceil(float64(i.config.ConcurrentScanLimit) / 2.0)),
			ConfigMapMounts: map[string]string{"perceptor-scanner": "/etc/perceptor_scanner"},
			Env: []envSecret{
				{
					EnvName:       i.config.HubUserPasswordEnvVar,
					SecretName:    i.config.ViperSecret,
					KeyFromSecret: "HubUserPassword",
				},
			},
			EmptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			Name:               i.config.ScannerImageName,
			Image:              paths["perceptor-scanner"],
			DockerSocket:       false,
			Port:               int32(i.config.ScannerPort),
			Cmd:                []string{"./perceptor-scanner"},
			Arg:                []floatstr.FloatOrString{i.GenerateArg("/etc/perceptor_scanner/perceptor_scanner.yaml", 0)},
			ServiceAccount:     i.config.ServiceAccounts["perceptor-image-facade"],
			ServiceAccountName: i.config.ServiceAccounts["perceptor-image-facade"],
			CPU:                defaultCPU,
			Memory:             defaultMem,
		},
		{
			ConfigMapMounts: map[string]string{"perceptor-imagefacade": "/etc/perceptor_imagefacade"},
			EmptyDirMounts: map[string]string{
				"var-images": "/var/images",
			},
			Name:               i.config.ImageFacadeImageName,
			Image:              paths["perceptor-imagefacade"],
			DockerSocket:       true,
			Port:               int32(i.config.ImageFacadePort),
			Cmd:                []string{"./perceptor-imagefacade"},
			Arg:                []floatstr.FloatOrString{i.GenerateArg("/etc/perceptor_imagefacade/perceptor_imagefacade.yaml", 0)},
			ServiceAccount:     i.config.ServiceAccounts["perceptor-image-facade"],
			ServiceAccountName: i.config.ServiceAccounts["perceptor-image-facade"],
			CPU:                defaultCPU,
			Memory:             defaultMem,
		},
	})

	// We dont create openshift perceivers if running kube... This needs to be avoided b/c the svc accounts
	// won't exist.
	if i.config.Openshift {
		i.AddReplicationControllerAndService([]*ReplicationController{
			{
				Replicas:        1,
				ConfigMapMounts: map[string]string{"perceiver": "/etc/perceiver"},
				EmptyDirMounts: map[string]string{
					"logs": "/tmp",
				},
				Name:               i.config.ImagePerceiverImageName,
				Image:              paths["image-perceiver"],
				Port:               int32(i.config.PerceiverPort),
				Cmd:                []string{"./image-perceiver"},
				Arg:                []floatstr.FloatOrString{i.GenerateArg("/etc/perceiver/perceiver.yaml", 0)},
				ServiceAccount:     i.config.ServiceAccounts["image-perceiver"],
				ServiceAccountName: i.config.ServiceAccounts["image-perceiver"],
				CPU:                defaultCPU,
				Memory:             defaultMem,
			},
		})
	}

	if i.config.PerceptorSkyfire {
		i.AddReplicationControllerAndService([]*ReplicationController{
			{
				Replicas:        1,
				ConfigMapMounts: map[string]string{"skyfire": "/etc/skyfire"},
				EmptyDirMounts: map[string]string{
					"logs": "/tmp",
				},
				Env: []envSecret{
					{
						EnvName:       i.config.HubUserPasswordEnvVar,
						SecretName:    i.config.ViperSecret,
						KeyFromSecret: "HubUserPassword",
					},
				},
				Name:               i.config.SkyfireImageName,
				Image:              paths["perceptor-skyfire"],
				Port:               3005,
				Cmd:                []string{"./skyfire"},
				Arg:                []floatstr.FloatOrString{i.GenerateArg("/etc/skyfire/skyfire.yaml", 0)},
				ServiceAccount:     i.config.ServiceAccounts["image-perceiver"],
				ServiceAccountName: i.config.ServiceAccounts["image-perceiver"],
				CPU:                defaultCPU,
				Memory:             defaultMem,
			},
		})
	}
}

func (i *Installer) deploy() {
	// Create all the resources.  Note that we'll panic after creating ANY
	// resource that fails.  Thats intentional

	// Create the configmaps first
	log.Println("Creating config maps : Dry Run ")
	for _, kconfigMap := range i.configMaps {
		wrapper := &types.ConfigMapWrapper{ConfigMap: *kconfigMap}
		configMap, err := converters.Convert_Koki_ConfigMap_to_Kube_v1_ConfigMap(wrapper)
		if err != nil {
			panic(err)
		}
		log.Println("*********************************************")
		log.Println("Creating config maps:", configMap)
		if !i.config.DryRun {
			log.Println("creating config map")
			_, err := i.client.Core().ConfigMaps(i.config.Namespace).Create(configMap)
			if err != nil {
				panic(err)
			}
		} else {
			i.prettyPrint(configMap)
		}
	}

	// Deploy the replication controllers
	for _, krc := range i.replicationControllers {
		wrapper := &types.ReplicationControllerWrapper{ReplicationController: *krc}
		rc, err := converters.Convert_Koki_ReplicationController_to_Kube_v1_ReplicationController(wrapper)
		if err != nil {
			panic(err)
		}
		i.prettyPrint(rc)
		if !i.config.DryRun {
			_, err := i.client.Core().ReplicationControllers(i.config.Namespace).Create(rc)
			if err != nil {
				panic(err)
			}
		}
	}

	// Deploy the pods
	for _, kpod := range i.pods {
		wrapper := &types.PodWrapper{Pod: *kpod}
		pod, err := converters.Convert_Koki_Pod_to_Kube_v1_Pod(wrapper)
		if err != nil {
			panic(err)
		}
		i.prettyPrint(pod)
		if !i.config.DryRun {
			_, err := i.client.Core().Pods(i.config.Namespace).Create(pod)
			if err != nil {
				panic(err)
			}
		}
	}

	// Deploy the services
	for _, ksvcI := range i.services {
		sWrapper := &types.ServiceWrapper{Service: *ksvcI}
		svcI, err := converters.Convert_Koki_Service_To_Kube_v1_Service(sWrapper)
		if err != nil {
			panic(err)
		}
		if i.config.DryRun {
			// service dont really need much debug...
			// i.prettyPrint(svcI)
		} else {
			_, err := i.client.Core().Services(i.config.Namespace).Create(svcI)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (i *Installer) sanityCheckServices() bool {
	isValid := func(cn string) bool {
		for _, valid := range []string{"perceptor", "pod-perceiver", "image-perceiver", "perceptor-scanner", "perceptor-image-facade"} {
			if cn == valid {
				return true
			}
		}
		return false
	}
	for cn := range i.config.ServiceAccounts {
		if !isValid(cn) {
			log.Print("[protoform] failed at verifiying that the container name for a svc account was valid!")
			log.Fatalln(cn)
		}
	}
	return true
}

func (i *Installer) createConfigMaps() {
	for k, v := range i.config.toMap() {
		mapConfig := api.ConfigMapConfig{
			Name:      k,
			Namespace: i.config.Namespace,
			Data:      v,
		}
		i.AddConfigMap(mapConfig)
	}
}

func (i *Installer) generateContainerPaths() map[string]string {
	config := i.config
	return map[string]string{
		"perceptor":             fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.PerceptorImageName, config.PerceptorContainerVersion),
		"perceptor-scanner":     fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ScannerImageName, config.ScannerContainerVersion),
		"pod-perceiver":         fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.PodPerceiverImageName, config.PerceiverContainerVersion),
		"image-perceiver":       fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ImagePerceiverImageName, config.PerceiverContainerVersion),
		"perceptor-imagefacade": fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.ImageFacadeImageName, config.ImageFacadeContainerVersion),
		"perceptor-skyfire":     fmt.Sprintf("%s/%s/%s:%s", config.Registry, config.ImagePath, config.SkyfireImageName, config.SkyfireContainerVersion),
	}
}

func (i *Installer) prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

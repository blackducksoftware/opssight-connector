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

package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	hub_v2 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssight_v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// CreateContainer will create the container
func CreateContainer(config *horizonapi.ContainerConfig, envs []*horizonapi.EnvConfig, volumeMounts []*horizonapi.VolumeMountConfig, ports []*horizonapi.PortConfig,
	actionConfig *horizonapi.ActionConfig, livenessProbeConfigs []*horizonapi.ProbeConfig, readinessProbeConfigs []*horizonapi.ProbeConfig) *components.Container {

	container := components.NewContainer(*config)

	for _, env := range envs {
		container.AddEnv(*env)
	}

	for _, volumeMount := range volumeMounts {
		container.AddVolumeMount(*volumeMount)
	}

	for _, port := range ports {
		container.AddPort(*port)
	}

	if actionConfig != nil {
		container.AddPostStartAction(*actionConfig)
	}

	for _, livenessProbe := range livenessProbeConfigs {
		container.AddLivenessProbe(*livenessProbe)
	}

	for _, readinessProbe := range readinessProbeConfigs {
		container.AddReadinessProbe(*readinessProbe)
	}

	return container
}

// CreateGCEPersistentDiskVolume will create a GCE Persistent disk volume for a pod
func CreateGCEPersistentDiskVolume(volumeName string, diskName string, fsType string) *components.Volume {
	gcePersistentDiskVol := components.NewGCEPersistentDiskVolume(horizonapi.GCEPersistentDiskVolumeConfig{
		VolumeName: volumeName,
		DiskName:   diskName,
		FSType:     fsType,
	})

	return gcePersistentDiskVol
}

// CreateEmptyDirVolumeWithoutSizeLimit will create a empty directory for a pod
func CreateEmptyDirVolumeWithoutSizeLimit(volumeName string) (*components.Volume, error) {
	emptyDirVol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: volumeName,
	})

	return emptyDirVol, err
}

// CreatePersistentVolumeClaimVolume will create a PVC claim for a pod
func CreatePersistentVolumeClaimVolume(volumeName string, pvcName string) (*components.Volume, error) {
	pvcVol := components.NewPVCVolume(horizonapi.PVCVolumeConfig{
		PVCName:    pvcName,
		VolumeName: volumeName,
	})

	return pvcVol, nil
}

// CreateEmptyDirVolume will create a empty directory for a pod
func CreateEmptyDirVolume(volumeName string, sizeLimit string) (*components.Volume, error) {
	emptyDirVol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: volumeName,
		SizeLimit:  sizeLimit,
	})

	return emptyDirVol, err
}

// CreateConfigMapVolume will mount the config map for a pod
func CreateConfigMapVolume(volumeName string, mapName string, defaultMode int) (*components.Volume, error) {
	configMapVol := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		DefaultMode:     IntToInt32(defaultMode),
		MapOrSecretName: mapName,
	})

	return configMapVol, nil
}

// CreateSecretVolume will mount the secret for a pod
func CreateSecretVolume(volumeName string, secretName string, defaultMode int) (*components.Volume, error) {
	secretVol := components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		DefaultMode:     IntToInt32(defaultMode),
		MapOrSecretName: secretName,
	})

	return secretVol, nil
}

// CreatePod will create the pod
func CreatePod(name string, serviceAccount string, volumes []*components.Volume, containers []*Container, initContainers []*Container, affinityConfigs []horizonapi.AffinityConfig) *components.Pod {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: name,
	})

	if !strings.EqualFold(serviceAccount, "") {
		pod.GetObj().Account = serviceAccount
	}

	for _, volume := range volumes {
		pod.AddVolume(volume)
	}

	pod.AddLabels(map[string]string{
		"app":  name,
		"tier": name,
	})

	for _, affinityConfig := range affinityConfigs {
		pod.AddAffinity(affinityConfig)
	}

	for _, containerConfig := range containers {
		container := CreateContainer(containerConfig.ContainerConfig, containerConfig.EnvConfigs, containerConfig.VolumeMounts, containerConfig.PortConfig,
			containerConfig.ActionConfig, containerConfig.LivenessProbeConfigs, containerConfig.ReadinessProbeConfigs)
		pod.AddContainer(container)
	}

	for _, initContainerConfig := range initContainers {
		initContainer := CreateContainer(initContainerConfig.ContainerConfig, initContainerConfig.EnvConfigs, initContainerConfig.VolumeMounts,
			initContainerConfig.PortConfig, initContainerConfig.ActionConfig, initContainerConfig.LivenessProbeConfigs, initContainerConfig.ReadinessProbeConfigs)
		err := pod.AddInitContainer(initContainer)
		if err != nil {
			log.Printf("failed to create the init container because %+v", err)
		}
	}

	return pod
}

// CreateDeployment will create a deployment
func CreateDeployment(deploymentConfig *horizonapi.DeploymentConfig, pod *components.Pod) *components.Deployment {
	deployment := components.NewDeployment(*deploymentConfig)

	deployment.AddMatchLabelsSelectors(map[string]string{
		"app":  deploymentConfig.Name,
		"tier": deploymentConfig.Name,
	})
	deployment.AddPod(pod)
	return deployment
}

// CreateDeploymentFromContainer will create a deployment with multiple containers inside a pod
func CreateDeploymentFromContainer(deploymentConfig *horizonapi.DeploymentConfig, serviceAccount string, containers []*Container, volumes []*components.Volume, initContainers []*Container, affinityConfigs []horizonapi.AffinityConfig) *components.Deployment {
	pod := CreatePod(deploymentConfig.Name, serviceAccount, volumes, containers, initContainers, affinityConfigs)
	deployment := CreateDeployment(deploymentConfig, pod)
	return deployment
}

// CreateReplicationController will create a replication controller
func CreateReplicationController(replicationControllerConfig *horizonapi.ReplicationControllerConfig, pod *components.Pod) *components.ReplicationController {
	rc := components.NewReplicationController(*replicationControllerConfig)
	rc.AddLabelSelectors(map[string]string{
		"app":  replicationControllerConfig.Name,
		"tier": replicationControllerConfig.Name,
	})
	rc.AddPod(pod)
	return rc
}

// CreateReplicationControllerFromContainer will create a replication controller with multiple containers inside a pod
func CreateReplicationControllerFromContainer(replicationControllerConfig *horizonapi.ReplicationControllerConfig, serviceAccount string, containers []*Container, volumes []*components.Volume, initContainers []*Container, affinityConfigs []horizonapi.AffinityConfig) *components.ReplicationController {
	pod := CreatePod(replicationControllerConfig.Name, serviceAccount, volumes, containers, initContainers, affinityConfigs)
	rc := CreateReplicationController(replicationControllerConfig, pod)
	return rc
}

// CreateService will create the service
func CreateService(name string, label string, namespace string, port string, target string, serviceType horizonapi.ClusterIPServiceType) *components.Service {
	svcConfig := horizonapi.ServiceConfig{
		Name:          name,
		Namespace:     namespace,
		IPServiceType: serviceType,
	}

	mySvc := components.NewService(svcConfig)
	portVal, _ := strconv.Atoi(port)
	myPort := &horizonapi.ServicePortConfig{
		Name:       fmt.Sprintf("port-" + name),
		Port:       int32(portVal),
		TargetPort: target,
		Protocol:   horizonapi.ProtocolTCP,
	}

	mySvc.AddPort(*myPort)
	mySvc.AddSelectors(map[string]string{"app": label})

	return mySvc
}

// CreateServiceWithMultiplePort will create the service with multiple port
func CreateServiceWithMultiplePort(name string, label string, namespace string, ports []string, serviceType horizonapi.ClusterIPServiceType) *components.Service {
	svcConfig := horizonapi.ServiceConfig{
		Name:          name,
		Namespace:     namespace,
		IPServiceType: serviceType,
	}

	mySvc := components.NewService(svcConfig)

	for _, port := range ports {
		portVal, _ := strconv.Atoi(port)
		myPort := &horizonapi.ServicePortConfig{
			Name:       fmt.Sprintf("port-" + port),
			Port:       int32(portVal),
			TargetPort: port,
			Protocol:   horizonapi.ProtocolTCP,
		}
		mySvc.AddPort(*myPort)
	}

	mySvc.AddSelectors(map[string]string{"app": label})

	return mySvc
}

// CreateSecretFromFile will create the secret from file
func CreateSecretFromFile(clientset *kubernetes.Clientset, jsonFile string, namespace string, name string, dataKey string) (*v1.Secret, error) {
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Panicf("Unable to read the secret file %s due to error: %v\n", jsonFile, err)
	}

	return clientset.CoreV1().Secrets(namespace).Create(&v1.Secret{
		Type:       v1.SecretTypeOpaque,
		StringData: map[string]string{dataKey: string(file)},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// CreateSecret will create the secret
func CreateSecret(clientset *kubernetes.Clientset, namespace string, name string, stringData map[string]string) (*v1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Create(&v1.Secret{
		Type:       v1.SecretTypeOpaque,
		StringData: stringData,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

// GetSecret will create the secret
func GetSecret(clientset *kubernetes.Clientset, namespace string, name string) (*v1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

// ReadFromFile will read the file
func ReadFromFile(filePath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filePath)
	return file, err
}

// GetConfigMap will get the config map
func GetConfigMap(clientset *kubernetes.Clientset, namespace string, name string) (*v1.ConfigMap, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

// UpdateConfigMap updates a config map
func UpdateConfigMap(clientset *kubernetes.Clientset, namespace string, configMap *v1.ConfigMap) error {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Update(configMap)
	return err
}

// // CreateSecret will create the secret
// func CreateNewSecret(secretConfig *horizonapi.SecretConfig) *components.Secret {
//
// 	secret := components.NewSecret(horizonapi.SecretConfig{Namespace: secretConfig.Namespace, Name: secretConfig.Name, Type: secretConfig.Type})
//
// 	secret.AddData(secretConfig.Data)
// 	secret.AddStringData(secretConfig.StringData)
// 	secret.AddLabels(secretConfig.Labels)
// 	secret.AddAnnotations(secretConfig.Annotations)
//
// 	return secret
// }
//
// // CreateConfigMap will create the configMap
// func CreateConfigMap(configMapConfig *horizonapi.ConfigMapConfig) *components.ConfigMap {
//
// 	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: configMapConfig.Namespace, Name: configMapConfig.Name})
//
// 	configMap.AddData(configMapConfig.Data)
// 	configMap.AddLabels(configMapConfig.Labels)
// 	configMap.AddAnnotations(configMapConfig.Annotations)
//
// 	return configMap
// }

// CreateNamespace will create the namespace
func CreateNamespace(clientset *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      namespace,
		},
	})
}

// GetNamespace will get the namespace
func GetNamespace(clientset *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
}

// DeleteNamespace will delete the namespace
func DeleteNamespace(clientset *kubernetes.Clientset, namespace string) error {
	return clientset.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
}

// GetPods will get the input pods corresponding to a namespace
func GetPods(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.Pod, error) {
	return clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

// GetAllPodsForNamespace will get all the pods corresponding to a namespace
func GetAllPodsForNamespace(clientset *kubernetes.Clientset, namespace string) (*corev1.PodList, error) {
	return clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

// GetAllReplicationControllersForNamespace will get all the replication controllers corresponding to a namespace
func GetAllReplicationControllersForNamespace(clientset *kubernetes.Clientset, namespace string) (*corev1.ReplicationControllerList, error) {
	return clientset.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{})
}

// GetAllDeploymentsForNamespace will get all the deployments corresponding to a namespace
func GetAllDeploymentsForNamespace(clientset *kubernetes.Clientset, namespace string) (*appsv1.DeploymentList, error) {
	return clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
}

// CreatePersistentVolume will create the persistent volume
func CreatePersistentVolume(clientset *kubernetes.Clientset, name string, storageClass string, claimSize string, nfsPath string, nfsServer string) (*corev1.PersistentVolume, error) {
	pvQuantity, _ := resource.ParseQuantity(claimSize)
	return clientset.CoreV1().PersistentVolumes().Create(&corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name,
			Name:      name,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity:         map[corev1.ResourceName]resource.Quantity{corev1.ResourceStorage: pvQuantity},
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: storageClass,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Path:   nfsPath,
					Server: nfsServer,
				},
			},
		},
	})
}

// DeletePersistentVolume will delete the persistent volume
func DeletePersistentVolume(clientset *kubernetes.Clientset, name string) error {
	return clientset.CoreV1().PersistentVolumes().Delete(name, &metav1.DeleteOptions{})
}

// CreatePersistentVolumeClaim will create the persistent volume claim
func CreatePersistentVolumeClaim(name string, namespace string, pvcClaimSize string, storageClass string, accessMode horizonapi.PVCAccessModeType) (*components.PersistentVolumeClaim, error) {

	// Workaround so that storageClass does not get set to "", which prevent Kube from using the default storageClass
	var class *string
	if len(storageClass) == 0 {
		class = nil
	} else {
		class = &storageClass
	}

	postgresPVC, err := components.NewPersistentVolumeClaim(horizonapi.PVCConfig{
		Name:      name,
		Namespace: namespace,
		// VolumeName: createHub.Name,
		Size:  pvcClaimSize,
		Class: class,
	})
	if err != nil {
		return nil, err
	}
	postgresPVC.AddAccessMode(accessMode)

	return postgresPVC, nil
}

// ValidateServiceEndpoint will validate whether the service endpoint is ready to serve
func ValidateServiceEndpoint(clientset *kubernetes.Clientset, namespace string, name string) (*v1.Endpoints, error) {
	var endpoint *v1.Endpoints
	var err error
	for i := 0; i < 20; i++ {
		endpoint, err = GetServiceEndPoints(clientset, namespace, name)
		if err != nil {
			log.Infof("waiting for %s endpoint in %s", name, namespace)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	return endpoint, err
}

// WaitForServiceEndpointReady will wait for the service endpoint to start the service
func WaitForServiceEndpointReady(clientset *kubernetes.Clientset, namespace string, name string) error {
	endpoint, err := ValidateServiceEndpoint(clientset, namespace, name)
	if err != nil {
		return fmt.Errorf("unable to get service endpoint %s in %s because %+v", name, namespace, err)
	}
	for _, subset := range endpoint.Subsets {
		if len(subset.NotReadyAddresses) > 0 {
			for {
				log.Infof("waiting for %s in %s to be cloned/backed up", name, namespace)
				svc, err := GetServiceEndPoints(clientset, namespace, name)
				if err != nil {
					return fmt.Errorf("unable to get service endpoint %s in %s because %+v", name, namespace, err)
				}

				for _, subset := range svc.Subsets {
					if len(subset.Addresses) > 0 {
						return nil
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}
	return nil
}

// ValidatePodsAreRunningInNamespace will validate whether the pods are running in a given namespace
func ValidatePodsAreRunningInNamespace(clientset *kubernetes.Clientset, namespace string) error {
	pods, err := GetAllPodsForNamespace(clientset, namespace)
	if err != nil {
		return fmt.Errorf("unable to list the pods in namespace %s due to %+v", namespace, err)
	}

	allPodExist := ValidatePodsAreRunning(clientset, pods)
	if !allPodExist {
		ValidatePodsAreRunningInNamespace(clientset, namespace)
	}
	return nil
}

// ValidatePodsAreRunning will validate whether the pods are running
func ValidatePodsAreRunning(clientset *kubernetes.Clientset, pods *corev1.PodList) bool {
	// Check whether all pods are running
	for _, podList := range pods.Items {
		for {
			pod, _ := clientset.CoreV1().Pods(podList.Namespace).Get(podList.Name, metav1.GetOptions{})
			if strings.EqualFold(pod.Name, "") {
				log.Infof("pod %s is restarted in %s..... checking all pod status again...", podList.Name, podList.Namespace)
				return false
			}
			if strings.EqualFold(string(pod.Status.Phase), "Running") {
				break
			}
			log.Infof("pod %s is in %s status... waiting 10 seconds", pod.Name, string(pod.Status.Phase))
			time.Sleep(10 * time.Second)
		}
	}
	return true
}

// FilterPodByNamePrefixInNamespace will filter the pod based on pod name prefix from a list a pods in a given namespace
func FilterPodByNamePrefixInNamespace(clientset *kubernetes.Clientset, namespace string, prefix string) (*corev1.Pod, error) {
	pods, err := GetAllPodsForNamespace(clientset, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to list the pods in namespace %s due to %+v", namespace, err)
	}

	pod := FilterPodByNamePrefix(pods, prefix)
	if pod != nil {
		return pod, nil
	}
	return nil, fmt.Errorf("unable to find the pod with prefix %s", prefix)
}

// FilterPodByNamePrefix will filter the pod based on pod name prefix from a list a pods
func FilterPodByNamePrefix(pods *corev1.PodList, prefix string) *corev1.Pod {
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			return &pod
		}
	}
	return nil
}

// CreateExecContainerRequest will create the request to exec into kubernetes pod
func CreateExecContainerRequest(clientset *kubernetes.Clientset, pod *corev1.Pod) *rest.Request {
	return clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", pod.Spec.Containers[0].Name).
		VersionedParams(&corev1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
}

// NewStringReader will convert string array to string reader object
func NewStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}

// NewKubeClientFromOutsideCluster will get the kube Configuration from outside the cluster
func newKubeClientFromOutsideCluster() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Errorf("error creating default client config: %s", err)
		return nil, err
	}
	return config, err
}

// GetKubeConfig will get the kube configuration
func GetKubeConfig() (*rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Infof("unable to get in cluster config due to %v", err)
		log.Infof("trying to use local config")
		config, err = newKubeClientFromOutsideCluster()
		if err != nil {
			log.Errorf("unable to retrive the local config due to %v", err)
			log.Panicf("failed to find a valid cluster config")
		}
	}

	return config, err
}

// GetService will get the service information for the input service name inside the input namespace
func GetService(clientset *kubernetes.Clientset, namespace string, serviceName string) (*v1.Service, error) {
	return clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
}

// GetServiceEndPoints will get the service endpoint information for the input service name inside the input namespace
func GetServiceEndPoints(clientset *kubernetes.Clientset, namespace string, serviceName string) (*v1.Endpoints, error) {
	return clientset.CoreV1().Endpoints(namespace).Get(serviceName, metav1.GetOptions{})
}

// ListStorageClass will list all the storageClass in the cluster
func ListStorageClass(clientset *kubernetes.Clientset) (*v1beta1.StorageClassList, error) {
	return clientset.StorageV1beta1().StorageClasses().List(metav1.ListOptions{})
}

// GetPVC will get the PVC for the given name
func GetPVC(clientset *kubernetes.Clientset, namespace string, name string) (*v1.PersistentVolumeClaim, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
}

// ListOpsSights will list all opssights in the cluster
func ListOpsSights(opssightClientset *opssightclientset.Clientset, namespace string) (*opssight_v1.OpsSightList, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).List(metav1.ListOptions{})
}

// GetOpsSight will get OpsSight in the cluster
func GetOpsSight(opssightClientset *opssightclientset.Clientset, namespace string, name string) (*opssight_v1.OpsSight, error) {
	return opssightClientset.SynopsysV1().OpsSights(namespace).Get(name, metav1.GetOptions{})
}

// GetOpsSights gets all opssights
func GetOpsSights(clientSet *opssightclientset.Clientset) (*opssight_v1.OpsSightList, error) {
	return clientSet.SynopsysV1().OpsSights(metav1.NamespaceAll).List(metav1.ListOptions{})
}

// ListHubs will list all hubs in the cluster
func ListHubs(hubClientset *hubclientset.Clientset, namespace string) (*hub_v2.BlackduckList, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).List(metav1.ListOptions{})
}

// WatchHubs will watch for hub events in the cluster
func WatchHubs(hubClientset *hubclientset.Clientset, namespace string) (watch.Interface, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Watch(metav1.ListOptions{})
}

// CreateHub will create hub in the cluster
func CreateHub(hubClientset *hubclientset.Clientset, namespace string, createHub *hub_v2.Blackduck) (*hub_v2.Blackduck, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Create(createHub)
}

// GetHub will get hubs in the cluster
func GetHub(hubClientset *hubclientset.Clientset, namespace string, name string) (*hub_v2.Blackduck, error) {
	return hubClientset.SynopsysV1().Blackducks(namespace).Get(name, metav1.GetOptions{})
}

// ListHubPV will list all the persistent volumes attached to each hub in the cluster
func ListHubPV(hubClientset *hubclientset.Clientset, namespace string) (map[string]string, error) {
	var pvList map[string]string
	pvList = make(map[string]string)
	hubs, err := ListHubs(hubClientset, namespace)
	if err != nil {
		log.Errorf("unable to list the hubs due to %+v", err)
		return pvList, err
	}
	for _, hub := range hubs.Items {
		if hub.Spec.PersistentStorage {
			pvList[hub.Name] = fmt.Sprintf("%s (%s)", hub.Name, hub.Status.PVCVolumeName["blackduck-postgres"])
		}
	}
	return pvList, nil
}

// IntToInt32 will convert from int to int32
func IntToInt32(i int) *int32 {
	j := int32(i)
	return &j
}

// IntToInt64 will convert from int to int64
func IntToInt64(i int) *int64 {
	j := int64(i)
	return &j
}

func getBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// RandomString will generate the random string
func RandomString(n int) (string, error) {
	b, err := getBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// CreateServiceAccount creates a service account
func CreateServiceAccount(namespace string, name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      name,
		Namespace: namespace,
	})

	return serviceAccount
}

// CreateClusterRoleBinding creates a cluster role binding
func CreateClusterRoleBinding(namespace string, name string, serviceAccountName string, clusterRoleAPIGroup string, clusterRoleKind string, clusterRoleName string) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       name,
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      serviceAccountName,
		Namespace: namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: clusterRoleAPIGroup,
		Kind:     clusterRoleKind,
		Name:     clusterRoleName,
	})

	return clusterRoleBinding
}

// DeleteClusterRoleBinding delete a cluster role binding
func DeleteClusterRoleBinding(clientset *kubernetes.Clientset, name string) error {
	return clientset.Rbac().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
}

// DeleteClusterRole delete a cluster role binding
func DeleteClusterRole(clientset *kubernetes.Clientset, name string) error {
	return clientset.Rbac().ClusterRoles().Delete(name, &metav1.DeleteOptions{})
}

// GetOpenShiftRoutes get a OpenShift routes
func GetOpenShiftRoutes(routeClient *routeclient.RouteV1Client, namespace string, name string) (*routev1.Route, error) {
	return routeClient.Routes(namespace).Get(name, metav1.GetOptions{})
}

// CreateOpenShiftRoutes creates a OpenShift routes
func CreateOpenShiftRoutes(routeClient *routeclient.RouteV1Client, namespace string, name string, routeKind string, serviceName string) (*routev1.Route, error) {
	return routeClient.Routes(namespace).Create(&routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: routev1.RouteSpec{
			TLS: &routev1.TLSConfig{Termination: routev1.TLSTerminationPassthrough},
			To: routev1.RouteTargetReference{
				Kind: routeKind,
				Name: serviceName,
			},
			Port: &routev1.RoutePort{TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: fmt.Sprintf("port-%s", serviceName)}},
		},
	})
}

// GetOpenShiftSecurityConstraint get a OpenShift security constraints
func GetOpenShiftSecurityConstraint(osSecurityClient *securityclient.SecurityV1Client, name string) (*securityv1.SecurityContextConstraints, error) {
	return osSecurityClient.SecurityContextConstraints().Get(name, metav1.GetOptions{})
}

// UpdateOpenShiftSecurityConstraint updates a OpenShift security constraints
func UpdateOpenShiftSecurityConstraint(osSecurityClient *securityclient.SecurityV1Client, serviceAccounts []string, name string) error {
	scc, err := GetOpenShiftSecurityConstraint(osSecurityClient, name)
	if err != nil {
		return fmt.Errorf("failed to get scc %s: %v", name, err)
	}

	newUsers := []string{}
	// Only add the service account if it isn't already in the list of users for the privileged scc
	for _, sa := range serviceAccounts {
		exist := false
		for _, user := range scc.Users {
			if strings.Compare(user, sa) == 0 {
				exist = true
				break
			}
		}

		if !exist {
			newUsers = append(newUsers, sa)
		}
	}

	if len(newUsers) > 0 {
		scc.Users = append(scc.Users, newUsers...)

		_, err = osSecurityClient.SecurityContextConstraints().Update(scc)
		if err != nil {
			return fmt.Errorf("failed to update scc %s: %v", name, err)
		}
	}
	return err
}

// PatchReplicationController patch a replication controller
func PatchReplicationController(clientset *kubernetes.Clientset, old corev1.ReplicationController, new corev1.ReplicationController) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.ReplicationController{})
	if err != nil {
		return
	}
	_, err = clientset.CoreV1().ReplicationControllers(new.Namespace).Patch(new.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return
	}
}

// UniqueValues returns a unique subset of the string slice provided.
func UniqueValues(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

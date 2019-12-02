package util

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// HTTPGet returns the http response for the api
func HTTPGet(url string) (content []byte, err error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		proxyURL, _ := http.ProxyFromEnvironment(response.Request)
		if proxyURL != nil {
			return nil, fmt.Errorf("failed to fetch %s using proxy %s | %s", response.Request.URL.String(), proxyURL.String(), response.Status)
		}
		return nil, fmt.Errorf("failed to fetch %s | %s", response.Request.URL.String(), response.Status)
	}
	return ioutil.ReadAll(response.Body)
}

// ConvertYamlFileToRuntimeObjects converts the yaml file string to map of runtime object
func ConvertYamlFileToRuntimeObjects(stringContent string) (map[string]runtime.Object, error) {
	routev1.AddToScheme(scheme.Scheme)
	securityv1.AddToScheme(scheme.Scheme)

	listOfSingleK8sResourceYaml := strings.Split(stringContent, "---")
	mapOfUniqueIDToDesiredRuntimeObject := make(map[string]runtime.Object, 0)

	for _, singleYaml := range listOfSingleK8sResourceYaml {
		if singleYaml == "\n" || singleYaml == "" {
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		runtimeObject, groupVersionKind, err := decode([]byte(singleYaml), nil, nil)
		if err != nil {
			return nil, err
		}

		accessor := meta.NewAccessor()
		runtimeObjectKind := groupVersionKind.Kind
		runtimeObjectName, err := accessor.Name(runtimeObject)
		if err != nil {
			return nil, err
		}
		uniqueID := fmt.Sprintf("%s.%s", runtimeObjectKind, runtimeObjectName)
		mapOfUniqueIDToDesiredRuntimeObject[uniqueID] = runtimeObject
	}
	return mapOfUniqueIDToDesiredRuntimeObject, nil
}

// GetBaseYaml returns the base yaml as string for the given app and version
func GetBaseYaml(baseurl string, appName string, version string, fileName string) (string, error) {
	// only fetch the location of the latest if the version in the spec is not given
	url, err := url.Parse(baseurl)
	if err != nil {
		return "", err
	}

	url.Path = path.Join(url.Path, appName, version, fileName)

	return downloadAndConvertYamlToByteArray(url.String())
}

func downloadAndConvertYamlToByteArray(url string) (string, error) {
	versionBaseYamlAsByteArray, err := HTTPGet(url)
	if err != nil {
		return "", err
	}
	return string(versionBaseYamlAsByteArray), nil
}

// EncodeStringToBase64 will return encoded string to base64
func EncodeStringToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// UpdateRegistry updates the registry for all runtime objects containers
func UpdateRegistry(obj map[string]runtime.Object, registry string) (map[string]runtime.Object, error) {
	for _, v := range obj {
		if podspec := findPodSpec(reflect.ValueOf(v)); podspec != nil {
			if err := updateContainersImage(*podspec, registry); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func findPodSpec(t reflect.Value) *corev1.PodSpec {
	podSpecType := reflect.TypeOf(corev1.PodSpec{})

	switch t.Kind() {
	case reflect.Ptr:
		return findPodSpec(t.Elem())
	case reflect.Struct:
		if t.Type() == podSpecType && t.CanInterface() {
			podSpec, _ := t.Interface().(corev1.PodSpec)
			return &podSpec
		}
		for i := 0; i < t.NumField(); i++ {
			if podSpec := findPodSpec(t.Field(i)); podSpec != nil {
				return podSpec
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			if podSpec := findPodSpec(t.Index(i)); podSpec != nil {
				return podSpec
			}
		}
	case reflect.Map:
		for _, key := range t.MapKeys() {
			if podSpec := findPodSpec(t.MapIndex(key)); podSpec != nil {
				return podSpec
			}
		}
	}

	return nil
}

func updateContainersImage(podSpec corev1.PodSpec, registry string) error {
	for containerIndex, container := range podSpec.Containers {
		newImage, err := generateNewImage(container.Image, registry)
		if err != nil {
			return err
		}
		podSpec.Containers[containerIndex].Image = newImage
	}

	for initContainerIndex, initContainer := range podSpec.InitContainers {
		newImage, err := generateNewImage(initContainer.Image, registry)
		if err != nil {
			return err
		}
		podSpec.InitContainers[initContainerIndex].Image = newImage
	}
	return nil
}

func generateNewImage(currentImage string, registry string) (string, error) {
	imageTag, err := getImageAndTag(currentImage)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", registry, imageTag), nil
}

func getImageAndTag(image string) (string, error) {
	r := regexp.MustCompile(`^(|.*/)([a-zA-Z_0-9-.:]+)$`)
	groups := r.FindStringSubmatch(image)
	if len(groups) < 3 && len(groups[2]) == 0 {
		return "", fmt.Errorf("couldn't find image and tags in [%s]", image)
	}
	return groups[2], nil
}

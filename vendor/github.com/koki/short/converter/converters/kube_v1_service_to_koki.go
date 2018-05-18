package converters

import (
	"net"
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/koki/short/types"
	"github.com/koki/short/util/intbool"
	serrors "github.com/koki/structurederrors"
)

func Convert_Kube_v1_Service_to_Koki_Service(kubeService *v1.Service) (*types.ServiceWrapper, error) {
	var err error
	kokiWrapper := &types.ServiceWrapper{}
	kokiService := &kokiWrapper.Service

	kokiService.Name = kubeService.Name
	kokiService.Namespace = kubeService.Namespace
	kokiService.Version = kubeService.APIVersion
	kokiService.Cluster = kubeService.ClusterName
	kokiService.Labels = kubeService.Labels
	kokiService.Annotations = kubeService.Annotations

	if kubeService.Spec.Type == v1.ServiceTypeExternalName {
		kokiService.ExternalName = kubeService.Spec.ExternalName
		return kokiWrapper, nil
	}

	kokiService.Type, err = convertServiceType(kubeService.Spec.Type)
	if err != nil {
		return nil, err
	}

	kokiService.Selector = kubeService.Spec.Selector
	kokiService.ClusterIP = types.ClusterIP(kubeService.Spec.ClusterIP)
	kokiService.ExternalIPs = convertExternalIPs(kubeService.Spec.ExternalIPs)
	kokiService.ClientIPAffinity = convertSessionAffinity(&kubeService.Spec)
	kokiService.PublishNotReadyAddresses = kubeService.Spec.PublishNotReadyAddresses

	kokiService.ExternalTrafficPolicy, err = convertExternalTrafficPolicy(kubeService.Spec.ExternalTrafficPolicy)
	if err != nil {
		return nil, err
	}

	kokiPort, kokiNodePort, kokiPorts := convertPorts(kubeService.Spec.Ports)
	if err != nil {
		return nil, err
	}
	kokiService.Port = kokiPort
	kokiService.NodePort = kokiNodePort
	kokiService.Ports = kokiPorts

	if kubeService.Spec.Type == v1.ServiceTypeLoadBalancer {
		loadBalancer, err := convertLoadBalancer(kubeService)
		if err != nil {
			return nil, err
		}
		kokiService.SetLoadBalancer(loadBalancer)
	}

	return kokiWrapper, nil
}

func convertServiceType(kubeType v1.ServiceType) (types.ClusterIPServiceType, error) {
	if len(kubeType) == 0 {
		return "", nil
	}

	switch kubeType {
	case v1.ServiceTypeClusterIP:
		return types.ClusterIPServiceTypeDefault, nil
	case v1.ServiceTypeNodePort:
		return types.ClusterIPServiceTypeNodePort, nil
	case v1.ServiceTypeLoadBalancer:
		return types.ClusterIPServiceTypeLoadBalancer, nil
	default:
		return "", serrors.InvalidInstanceError(kubeType)
	}
}

func convertLoadBalancerIngress(kubeIngress []v1.LoadBalancerIngress) ([]types.LoadBalancerIngress, error) {
	if len(kubeIngress) == 0 {
		return nil, nil
	}

	kokiIngress := make([]types.LoadBalancerIngress, len(kubeIngress))
	for index, singleKubeIngress := range kubeIngress {
		if len(singleKubeIngress.IP) > 0 {
			ip := net.ParseIP(singleKubeIngress.IP)
			if ip == nil {
				return nil, serrors.InvalidInstanceErrorf(singleKubeIngress, "invalid IP")
			}

			kokiIngress[index] = types.LoadBalancerIngress{IP: ip}
		} else {
			kokiIngress[index] = types.LoadBalancerIngress{
				Hostname: singleKubeIngress.Hostname,
			}
		}
	}

	return kokiIngress, nil
}

func convertLoadBalancerSources(kubeSources []string) []types.CIDR {
	kokiCIDRs := make([]types.CIDR, len(kubeSources))
	for i, kubeSource := range kubeSources {
		kokiCIDRs[i] = types.CIDR(kubeSource)
	}

	return kokiCIDRs
}

func convertLoadBalancer(kubeService *v1.Service) (*types.LoadBalancer, error) {
	var err error
	kokiLB := &types.LoadBalancer{}
	kokiLB.IP = types.IPAddr(kubeService.Spec.LoadBalancerIP)
	kokiLB.Ingress, err = convertLoadBalancerIngress(kubeService.Status.LoadBalancer.Ingress)
	if err != nil {
		return nil, err
	}
	kokiLB.Allowed = convertLoadBalancerSources(kubeService.Spec.LoadBalancerSourceRanges)
	kokiLB.HealthCheckNodePort = kubeService.Spec.HealthCheckNodePort

	return kokiLB, nil
}

func convertPort(kubePort v1.ServicePort) (*types.ServicePort, int32) {
	kokiPort := &types.ServicePort{}
	kokiPort.Expose = kubePort.Port
	if !reflect.DeepEqual(kubePort.TargetPort, intstr.FromInt(0)) {
		kokiPort.PodPort = &kubePort.TargetPort
	}

	kokiPort.Protocol = convertProtocol(kubePort.Protocol)

	return kokiPort, kubePort.NodePort
}

func convertPorts(kubePorts []v1.ServicePort) (*types.ServicePort, int32, []types.NamedServicePort) {
	if len(kubePorts) == 1 && len(kubePorts[0].Name) == 0 {
		// Just one port, and it's unnamed
		kokiPort, kokiNodePort := convertPort(kubePorts[0])
		return kokiPort, kokiNodePort, nil
	}

	kokiPorts := make([]types.NamedServicePort, len(kubePorts))
	for i, kubePort := range kubePorts {
		kokiPort, kokiNodePort := convertPort(kubePort)

		kokiPorts[i] = types.NamedServicePort{
			Name:     kubePort.Name,
			Port:     *kokiPort,
			NodePort: kokiNodePort,
		}
	}

	return nil, 0, kokiPorts
}

func convertExternalTrafficPolicy(kubePolicy v1.ServiceExternalTrafficPolicyType) (types.ExternalTrafficPolicy, error) {
	switch kubePolicy {
	case "":
		return types.ExternalTrafficPolicyNil, nil
	case v1.ServiceExternalTrafficPolicyTypeLocal:
		return types.ExternalTrafficPolicyLocal, nil
	case v1.ServiceExternalTrafficPolicyTypeCluster:
		return types.ExternalTrafficPolicyCluster, nil
	default:
		return "", serrors.InvalidInstanceError(kubePolicy)
	}
}

// Returns koki ClientIPAffinitySeconds.
func convertSessionAffinity(kubeSpec *v1.ServiceSpec) *intbool.IntOrBool {
	if kubeSpec.SessionAffinity == v1.ServiceAffinityClientIP {
		if kubeSpec.SessionAffinityConfig != nil && kubeSpec.SessionAffinityConfig.ClientIP != nil && kubeSpec.SessionAffinityConfig.ClientIP.TimeoutSeconds != nil {
			return intbool.FromInt(int(*kubeSpec.SessionAffinityConfig.ClientIP.TimeoutSeconds))
		}

		return intbool.FromBool(true)
	}

	return nil
}

func convertExternalIPs(kubeIPs []string) []types.IPAddr {
	kokiIPs := make([]types.IPAddr, len(kubeIPs))
	for i, kubeIP := range kubeIPs {
		kokiIPs[i] = types.IPAddr(kubeIP)
	}

	return kokiIPs
}

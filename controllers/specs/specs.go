package specs

import (
	"math/rand"

	configmaps "example.com/lb/apis/configmaps/v1alpha1"
	deployments "example.com/lb/apis/deployments/v1alpha1"
	services "example.com/lb/apis/services/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const numOfReplicas = 1
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$&*"

// BuildDeployment creates a kubernetes deployment specification
func BuildDeployment(ns string, deploy *deployments.Deployment) *appsv1.Deployment {
	name := "lb-deploy"
	replicas := int32(numOfReplicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: buildMetadata(name, ns, deploy.APIVersion, deploy.Kind, deploy.UID),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(name),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: buildMetadata(name, ns, deploy.APIVersion, deploy.Kind, deploy.UID),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "lb",
							// TODO allow choosing image
							Image: "traefik",
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: deploy.Spec.ContainerPort,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(deploy.Spec.MemoryLimit),
									v1.ResourceCPU:    resource.MustParse(deploy.Spec.CpuLimit),
								},
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(deploy.Spec.MemoryRequest),
									v1.ResourceCPU:    resource.MustParse(deploy.Spec.CpuRequest),
								},
							},
							Env: []v1.EnvVar{
								// TODO allow setting secrets and relevant secrets
								getEnvVarSecretSource("CERT", "cert", "key.crt"),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}
}

// BuildConfigMap will build a kubernetes config map for postgres
func BuildConfigMap(ns string, cm *configmaps.ConfigMap) *v1.ConfigMap {
	name := "lb-cm"
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "batch/v1/beta1",
		},
		ObjectMeta: buildMetadata(name, ns, cm.APIVersion, cm.Kind, cm.UID),
		Data: map[string]string{
			"static.yaml":  "####",
			"dynamic.yaml": "###",
		},
	}
}

// BuildService in kubernetes with pgDeploy port
func BuildService(ns string, service *services.Service) *v1.Service {
	name := "lb-service"
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(name, ns, service.APIVersion, service.Kind, service.UID),
		Spec: v1.ServiceSpec{
			Type:       v1.ServiceTypeLoadBalancer,
			Ports:      buildServicePorts(service),
			Selector:   buildLabels(name),
			IPFamilies: []v1.IPFamily{},
		},
	}
}

//
// ------------------------------------------------------------------------------------------------------- helpers -----------------------------------------------------------------------------
//

func buildServicePorts(serivce *services.Service) []v1.ServicePort {
	var ports []v1.ServicePort

	for _, s := range serivce.Spec.Ports {
		svc := v1.ServicePort{
			Name:     s.Name,
			Protocol: v1.Protocol(s.Protocol),
			Port:     s.Port,
		}
		ports = append(ports, svc)
	}

	return ports
}

func getEnvVarSecretSource(envName, name, secret string) v1.EnvVar {
	return v1.EnvVar{
		Name: envName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: name,
				},
				Key: secret,
			},
		},
	}
}

func buildMetadata(name, namespace, apiVersion, kind string, uid types.UID) metav1.ObjectMeta {
	controlled := true
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    buildLabels(name),
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: apiVersion,
				Kind:       kind,
				Name:       name,
				UID:        uid,
				Controller: &controlled,
			},
		},
	}
}

func buildLabels(name string) map[string]string {
	m := make(map[string]string)
	m["app"] = "lb"
	m["app.kubernetes.io/name"] = name
	m["app.kubernetes.io/component"] = name
	return m
}

func createSecret() map[string]string {
	m := make(map[string]string)
	m["secret"] = randStringBytes(12)
	return m
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

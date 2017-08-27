package k8s

import (
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func BuildClient() (*kubernetes.Clientset, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(k8sConfig)
}

func GetConfigMapData(namespace, configmap string) (map[string]string, error) {
	kc, err := BuildClient()
	if err != nil {
		return nil, err
	}
	cm, err := kc.ConfigMaps(namespace).Get(configmap)
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func UpdateReplicasCount(namespace, deployment string, desiredReplicas int) error {
	kc, err := BuildClient()
	if err != nil {
		return err
	}
	deployYaml, err := kc.Deployments(namespace).Get(deployment)
	if err != nil {
		return err
	}

	dp := int32(desiredReplicas)
	deployYaml.Spec.Replicas = &dp
	_, err = kc.Deployments(namespace).Update(deployYaml)
	return err
}

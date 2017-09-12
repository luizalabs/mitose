package k8s

import (
	"errors"
	"io/ioutil"

	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/watch"
	restclient "k8s.io/client-go/rest"
)

const namespaceSecret = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

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

func UpdateHPA(namespace, name string, min, max int) error {
	kc, err := BuildClient()
	if err != nil {
		return err
	}
	hpaYaml, err := kc.AutoscalingV1().HorizontalPodAutoscalers(namespace).Get(name)
	if err != nil {
		return err
	}
	min32 := int32(min)
	hpaYaml.Spec.MaxReplicas = int32(max)
	hpaYaml.Spec.MinReplicas = &min32

	_, err = kc.AutoscalingV1().HorizontalPodAutoscalers(namespace).Update(hpaYaml)
	return err
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

func WatchConfigMap(namespace string) (<-chan error, error) {
	kc, err := BuildClient()
	if err != nil {
		return nil, err
	}
	watcher, err := kc.ConfigMaps(namespace).Watch(k8sv1.ListOptions{})
	if err != nil {
		return nil, err
	}

	c := make(chan error)
	go func() {
		defer close(c)
		for e := range watcher.ResultChan() {
			if e.Type == watch.Error {
				c <- errors.New("error reading configmap")
			} else {
				c <- nil
			}
		}
	}()
	return c, nil
}

func GetCurrentNamespace() (string, error) {
	ns, err := ioutil.ReadFile(namespaceSecret)
	return string(ns), err
}

package k8s

import (
	"errors"
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
)

const namespaceSecret = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

var (
	ClientBuilder func() (kubernetes.Interface, error) = concretBuilder
	ReadFile      func(string) ([]byte, error)         = ioutil.ReadFile
)

func concretBuilder() (kubernetes.Interface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(k8sConfig)
}

func GetConfigMapData(namespace, configmap string) (map[string]string, error) {
	kc, err := ClientBuilder()
	if err != nil {
		return nil, err
	}
	cm, err := kc.CoreV1().ConfigMaps(namespace).Get(configmap, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm.Data, nil
}

func UpdateHPA(namespace, name string, min, max int) error {
	kc, err := ClientBuilder()
	if err != nil {
		return err
	}
	hpaYaml, err := kc.AutoscalingV1().
		HorizontalPodAutoscalers(namespace).
		Get(name, metav1.GetOptions{})
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
	kc, err := ClientBuilder()
	if err != nil {
		return err
	}
	deployYaml, err := kc.Extensions().
		Deployments(namespace).
		Get(deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}

	dp := int32(desiredReplicas)
	deployYaml.Spec.Replicas = &dp
	_, err = kc.Extensions().Deployments(namespace).Update(deployYaml)
	return err
}

func WatchConfigMap(namespace string) (<-chan error, error) {
	kc, err := ClientBuilder()
	if err != nil {
		return nil, err
	}
	watcher, err := kc.CoreV1().ConfigMaps(namespace).Watch(metav1.ListOptions{})
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

func UpdateConfigMap(namespace string, configmap map[string]string) error {
	kc, err := ClientBuilder()
	if err != nil {
		return err
	}
	_, err = kc.CoreV1().ConfigMaps(namespace).Update(&apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Data:       configmap,
	})
	return err
}

func GetCurrentNamespace() (string, error) {
	ns, err := ReadFile(namespaceSecret)
	return string(ns), err
}

package k8s

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	hpa_apisv1 "k8s.io/client-go/pkg/apis/autoscaling/v1"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var fakeK8sClient kubernetes.Interface

func fakeBuilder() (kubernetes.Interface, error) {
	return fakeK8sClient, nil
}

func TestGetCurrentNamespace(t *testing.T) {
	expectedNS := "fakeNS"
	fakeReadFile := func(string) ([]byte, error) {
		return []byte(expectedNS), nil
	}
	ReadFile = fakeReadFile

	actual, err := GetCurrentNamespace()
	if err != nil {
		t.Fatal("error getting current namespace", err)
	}
	if actual != expectedNS {
		t.Errorf("expected %s, got %s", expectedNS, actual)
	}
}

func TestK8sPkg(t *testing.T) {
	var testFuncs = []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{"TestGetConfigMapData", testGetConfigMapData},
		{"TestUpdateHPA", testUpdateHPA},
		{"TestUpdateReplicasCount", testUpdateReplicasCount},
		{"TestWatchConfigMap", testWatchConfigMap},
	}

	ClientBuilder = fakeBuilder

	for _, tf := range testFuncs {
		fakeK8sClient = fake.NewSimpleClientset()
		t.Run(tf.name, tf.testFunc)
	}

	ClientBuilder = concretBuilder
}

func testGetConfigMapData(t *testing.T) {
	fakeNS := "fakeNS"
	expectedKey := "fakeKey"
	expectedValue := "fakeValue"

	fakeCM := &v1.ConfigMap{Data: map[string]string{expectedKey: expectedValue}}
	fakeCM.Name = "fake"
	if _, err := fakeK8sClient.CoreV1().ConfigMaps(fakeNS).Create(fakeCM); err != nil {
		t.Fatal("error creating fake config map", err)
	}

	cm, err := GetConfigMapData(fakeNS, fakeCM.Name)
	if err != nil {
		t.Fatal("error getting config map", err)
	}
	if cm[expectedKey] != expectedValue {
		t.Errorf("expected %s, got %s", expectedValue, cm[expectedKey])
	}
}

func testUpdateHPA(t *testing.T) {
	fakeHPAName := "fakeHPA"
	fakeNS := "fakeNS"
	actualMin := int32(1)
	fakeHPA := &hpa_apisv1.HorizontalPodAutoscaler{
		Spec: hpa_apisv1.HorizontalPodAutoscalerSpec{
			MaxReplicas: 1,
			MinReplicas: &actualMin,
		},
	}
	fakeHPA.Name = fakeHPAName

	_, err := fakeK8sClient.AutoscalingV1().
		HorizontalPodAutoscalers(fakeNS).
		Create(fakeHPA)
	if err != nil {
		t.Fatal("error creating fake hpa", err)
	}

	expectedMaxReplicas := 2
	if err := UpdateHPA(fakeNS, fakeHPAName, 2, expectedMaxReplicas); err != nil {
		t.Fatal("error updating hpa", err)
	}

	hpa, err := fakeK8sClient.AutoscalingV1().
		HorizontalPodAutoscalers(fakeNS).
		Get(fakeHPAName, metav1.GetOptions{})
	if err != nil {
		t.Fatal("error getting fake hpa", err)
	}

	if hpa.Spec.MaxReplicas != int32(expectedMaxReplicas) {
		t.Errorf("expected %d, got %d", expectedMaxReplicas, hpa.Spec.MaxReplicas)
	}
}

func testUpdateReplicasCount(t *testing.T) {
	fakeNS := "fakeNS"
	fakeDeployName := "fakeDeploy"
	actualReplicas := int32(1)
	fakeDeploy := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{Replicas: &actualReplicas},
	}
	fakeDeploy.Name = fakeDeployName

	_, err := fakeK8sClient.Extensions().
		Deployments(fakeNS).
		Create(fakeDeploy)
	if err != nil {
		t.Fatal("error creating fake deploy", err)
	}

	expectedReplicas := 2
	if err := UpdateReplicasCount(fakeNS, fakeDeployName, expectedReplicas); err != nil {
		t.Fatal("error updating replicas of fake deploy", err)
	}

	deploy, err := fakeK8sClient.Extensions().
		Deployments(fakeNS).
		Get(fakeDeployName, metav1.GetOptions{})
	if err != nil {
		t.Fatal("error getting fake deploy", err)
	}

	if *deploy.Spec.Replicas != int32(expectedReplicas) {
		t.Errorf("expected %d, got %d", expectedReplicas, *deploy.Spec.Replicas)
	}
}

func testWatchConfigMap(t *testing.T) {
	ClientBuilder = fakeBuilder

	fakeNS := "fakeNS"
	fakeName := "fake"

	fakeCM := &v1.ConfigMap{Data: map[string]string{}}
	fakeCM.Name = fakeName
	if _, err := fakeK8sClient.CoreV1().ConfigMaps(fakeNS).Create(fakeCM); err != nil {
		t.Fatal("error creating fake config map", err)
	}

	if _, err := WatchConfigMap(fakeNS); err != nil {
		t.Fatal("error getting watcher to config map", err)
	}
}

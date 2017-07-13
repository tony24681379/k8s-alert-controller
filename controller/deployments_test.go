package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	testcore "k8s.io/client-go/testing"
)

func pod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: v1.NamespaceDefault,
		},
	}
}

func deployment() *v1beta1.Deployment {
	replicas := int32(3)
	dep := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: v1.NamespaceDefault,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: v1beta1.DeploymentStatus{
			AvailableReplicas: int32(0),
		},
	}
	return dep
}

func TestDeploymentUpdateReplicas(t *testing.T) {
	clientSet := fake.NewSimpleClientset(deployment())
	deployment := DeploymentController{Deployment: clientSet.Extensions().Deployments(v1.NamespaceDefault)}
	expectReplicas := int32(2)
	err := deployment.updateReplicas("foo", expectReplicas)

	if err != nil {
		t.Errorf("unexpected actions: %v", err)
	}

	actions := clientSet.Actions()
	if len(actions) != 2 {
		t.Errorf("unexpected actions: %+v, expected 2 actions (get, update)", actions)
	}
	if action, ok := actions[1].(testcore.UpdateAction); !ok || action.GetResource().GroupResource() != v1beta1.Resource("deployments") || *(action.GetObject().(*v1beta1.Deployment).Spec.Replicas) != expectReplicas {
		t.Errorf("unexpected action %v, expected update-deployment with replicas = %d", actions[1], expectReplicas)
	}
}

func TestDeploymentDeletePod(t *testing.T) {
	clientSet := fake.NewSimpleClientset(pod())
	deployment := DeploymentController{
		Pod: clientSet.CoreV1().Pods(v1.NamespaceDefault),
	}
	err := deployment.deletePod("foo")

	if err != nil {
		t.Errorf("unexpected actions: %v", err)
	}

	actions := clientSet.Actions()
	if len(actions) != 1 {
		t.Errorf("unexpected actions: %+v, expected 2 actions (get, update)", actions)
	}
	if action, ok := actions[0].(testcore.DeleteAction); !ok || action.GetResource().GroupResource() != v1.Resource("pods") {
		t.Errorf("unexpected action %v, expected delete-deployment ", actions[0])
	}
}

func TestWaitForAvailable(t *testing.T) {
	clientSet := fake.NewSimpleClientset(deployment())
	deployment := DeploymentController{
		Deployment: clientSet.Extensions().Deployments(v1.NamespaceDefault),
	}
	complete := make(chan bool)
	go func() {
		err := deployment.waitForAvailable("foo")
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		complete <- true
	}()
	replicas := int32(3)
	dep := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: v1.NamespaceDefault,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: v1beta1.DeploymentStatus{
			AvailableReplicas: int32(1),
		},
	}
	deployment.Deployment.Update(dep)
	<-complete
}

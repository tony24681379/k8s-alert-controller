package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	testcore "k8s.io/client-go/testing"
)

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
			AvailableReplicas: int32(1),
		},
	}
	return dep
}

func TestDeploymentUpdateReplicas(t *testing.T) {
	dep := deployment()
	clientSet := fake.NewSimpleClientset(dep)
	_, err := clientSet.Extensions().Deployments(v1.NamespaceDefault).Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	deployment := DeploymentController{Deployment: clientSet.Extensions().Deployments(v1.NamespaceDefault)}
	err = deployment.updateReplicas("foo", -1)

	if err != nil {
		t.Errorf("unexpected actions: %v", err)
	}

	actions := clientSet.Actions()
	if len(actions) != 3 {
		t.Errorf("unexpected actions: %+v, expected 2 actions (get, update)", actions)
	}
	expectReplicas := int32(2)
	if action, ok := actions[2].(testcore.UpdateAction); !ok || action.GetResource().GroupResource() != v1beta1.Resource("deployments") || *(action.GetObject().(*v1beta1.Deployment).Spec.Replicas) != expectReplicas {
		t.Errorf("unexpected action %v, expected update-deployment with replicas = %d", actions[1], expectReplicas)
	}
}

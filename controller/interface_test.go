package controller

import (
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

func TestContorllerFor(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	tests := []struct {
		resourceType   string
		namespace      string
		expectedResult controllerInterface
		expectedError  error
	}{
		{
			resourceType: DeploymentType,
			namespace:    v1.NamespaceDefault,
			expectedResult: &DeploymentController{
				Pod:        clientset.CoreV1().Pods(v1.NamespaceDefault),
				Deployment: clientset.Extensions().Deployments(v1.NamespaceDefault),
			},
			expectedError: nil,
		},
		{
			resourceType: DaemonSetType,
			namespace:    v1.NamespaceDefault,
			expectedResult: &DaemonSetController{
				Pod:       clientset.CoreV1().Pods(v1.NamespaceDefault),
				DaemonSet: clientset.Extensions().DaemonSets(v1.NamespaceDefault),
			},
			expectedError: nil,
		},
	}

	for i, tt := range tests {
		result, err := ControllerFor(clientset, tt.resourceType, tt.namespace)
		if !reflect.DeepEqual(result, tt.expectedResult) {
			t.Errorf("#%d: expected error=%s, get=%s", i, tt.expectedResult, result)
		}
		if err != nil {
			if err != tt.expectedError {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
	}
}

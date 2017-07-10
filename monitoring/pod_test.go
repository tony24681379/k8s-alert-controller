package monitoring

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestCheckResourceDeployment(t *testing.T) {
	tests := []struct {
		label          map[string]string
		deploymentList *v1beta1.DeploymentList
		expectedResult bool
		expectedName   string
		expectedError  error
	}{
		{
			label: map[string]string{
				"app": "test",
			},
			deploymentList: &v1beta1.DeploymentList{
				Items: []v1beta1.Deployment{
					v1beta1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1beta1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					},
				},
			},
			expectedResult: true,
			expectedName:   "foo",
			expectedError:  nil,
		},
		{
			label: map[string]string{
				"app": "test",
			},
			deploymentList: &v1beta1.DeploymentList{},
			expectedResult: false,
			expectedName:   "",
			expectedError:  nil,
		},
	}

	for i, tt := range tests {
		clientset := fake.NewSimpleClientset(tt.deploymentList)
		isDelpoyment, deploymentName, err := checkResourceDeployment(clientset, v1.NamespaceDefault, tt.label)
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
		if isDelpoyment != tt.expectedResult {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResult, isDelpoyment)
		}
		if deploymentName != tt.expectedName {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedName, deploymentName)
		}
	}
}

func TestCheckResourceDaemonSet(t *testing.T) {
	tests := []struct {
		label          map[string]string
		daemonSetList  *v1beta1.DaemonSetList
		expectedResult bool
		expectedName   string
		expectedError  error
	}{
		{
			label: map[string]string{
				"app": "test",
			},
			daemonSetList: &v1beta1.DaemonSetList{
				Items: []v1beta1.DaemonSet{
					v1beta1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1beta1.DaemonSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					},
				},
			},
			expectedResult: true,
			expectedName:   "foo",
			expectedError:  nil,
		},
		{
			label: map[string]string{
				"app": "test",
			},
			daemonSetList:  &v1beta1.DaemonSetList{},
			expectedResult: false,
			expectedName:   "",
			expectedError:  nil,
		},
	}

	for i, tt := range tests {
		clientset := fake.NewSimpleClientset(tt.daemonSetList)
		isDaemonSet, daemonSetName, err := checkResourceDaemonSet(clientset, v1.NamespaceDefault, tt.label)
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
		if isDaemonSet != tt.expectedResult {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResult, isDaemonSet)
		}
		if daemonSetName != tt.expectedName {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedName, daemonSetName)
		}
	}
}

func TestCheckResource(t *testing.T) {
	tests := []struct {
		podName              string
		label                map[string]string
		object               runtime.Object
		expectedResourceType string
		expectedResourceName string
		expectedError        error
	}{
		{
			podName: "foo",
			label: map[string]string{
				"app": "test",
			},
			object: &v1beta1.DaemonSetList{
				Items: []v1beta1.DaemonSet{
					v1beta1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1beta1.DaemonSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					},
				},
			},
			expectedResourceType: "daemonSet",
			expectedResourceName: "foo",
			expectedError:        nil,
		},
		{
			podName: "foo",
			label: map[string]string{
				"app": "test",
			},
			object: &v1beta1.DeploymentList{
				Items: []v1beta1.Deployment{
					v1beta1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1beta1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test",
								},
							},
						},
					},
				},
			},
			expectedResourceType: "deployment",
			expectedResourceName: "foo",
			expectedError:        nil,
		},
		{
			podName: "foo",
			label: map[string]string{
				"app": "test",
			},
			object:               &v1beta1.DaemonSetList{},
			expectedResourceType: "",
			expectedResourceName: "",
			expectedError:        fmt.Errorf("%s resource type not found", "foo"),
		},
	}

	for i, tt := range tests {
		clientset := fake.NewSimpleClientset(tt.object)
		resourceType, resourceName, err := CheckResource(clientset, tt.podName, v1.NamespaceDefault, tt.label)
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
		if resourceType != tt.expectedResourceType {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResourceType, resourceType)
		}
		if resourceName != tt.expectedResourceName {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResourceName, resourceName)
		}
	}
}

package controller

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

const (
	DeploymentType  = "deployment"
	StatefulSetType = "statefulSet"
	DaemonSetType   = "daemonSet"
)

type controllerInterface interface {
	RestartOnePod(resourceName, podName string) error
}

func ControllerFor(clientset *kubernetes.Clientset, resourceType string, namespace string) (controllerInterface, error) {
	switch resourceType {
	case DeploymentType:
		return &DeploymentController{
			Pod:        clientset.CoreV1().Pods(namespace),
			Deployment: clientset.Extensions().Deployments(namespace),
		}, nil
	case StatefulSetType:
		return nil, nil
	case DaemonSetType:
		return &DaemonSetController{
			Pod:       clientset.CoreV1().Pods(namespace),
			DaemonSet: clientset.Extensions().DaemonSets(namespace),
		}, nil
	}
	return nil, fmt.Errorf("no controller has been implemented for %q", resourceType)
}

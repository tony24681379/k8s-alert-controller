package monitoring

import (
	"sync"

	"fmt"

	"github.com/golang/glog"
	"github.com/tony24681379/k8s-alert-controller/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodRestart will restart the pod and return success string or error
func PodRestart(clientset *kubernetes.Clientset, podName, namespace string) (string, error) {
	glog.V(2).Info("PodRestart")
	glog.V(2).Info(namespace, podName)

	po, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err)
		return "", err
	}
	podLabel := po.GetLabels()
	glog.V(2).Info(podLabel)

	resourceType, resourceName, err := CheckResource(clientset, podName, namespace, podLabel)
	if err != nil {
		glog.Error(err)
		return "", err
	}

	controller, err := controller.ControllerFor(clientset, resourceType, namespace)
	if err != nil {
		glog.Error(err)
		return "", err
	}

	if err := controller.RestartOnePod(resourceName, podName); err != nil {
		glog.Error(err)
		return "", err
	}
	return resourceType + ": " + resourceName + " namespace:" + namespace, nil
}

// CheckResource return resourceType and resourceName
func CheckResource(clientset *kubernetes.Clientset, podName, namespace string, podLabel map[string]string) (resourceType string, resourceName string, err error) {
	var (
		isDelpoyment, isDaemonSet     bool
		deploymentName, daemonSetName string
		deploymentErr, daemonSetErr   error
		wg                            sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		isDelpoyment, deploymentName, deploymentErr = checkResourceDeployment(clientset, namespace, podLabel)
	}()

	go func() {
		defer wg.Done()
		isDaemonSet, daemonSetName, daemonSetErr = checkResourceDaemonSet(clientset, namespace, podLabel)
	}()

	wg.Wait()
	if deploymentErr != nil || daemonSetErr != nil {
		glog.Errorf("%s%s", deploymentErr, daemonSetErr)
		return "", "", fmt.Errorf("%s%s", deploymentErr, daemonSetErr)
	}

	if isDelpoyment {
		return controller.DeploymentType, deploymentName, nil
	} else if isDaemonSet {
		return controller.DaemonSetType, daemonSetName, nil
	}
	return "", "", fmt.Errorf("%s resource type not found", podName)
}

func checkResourceDeployment(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (bool, string, error) {
	deploymentList, err := clientset.Extensions().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return false, "", err
	}

	for _, d := range deploymentList.Items {
		selectorLabel := d.Spec.Selector.MatchLabels
		glog.V(2).Info(selectorLabel)
		equal := true
		for k, v := range selectorLabel {
			if podLabel[k] != v {
				equal = false
				break
			}
		}
		if equal {
			glog.Infof("resource %v %v is %v", d.GetNamespace(), d.GetName(), "deployment")
			return true, d.GetName(), nil
		}
	}
	return false, "", nil
}

func checkResourceDaemonSet(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (bool, string, error) {
	daemonSetList, err := clientset.Extensions().DaemonSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return false, "", err
	}

	for _, d := range daemonSetList.Items {
		selectorLabel := d.Spec.Selector.MatchLabels
		glog.V(2).Info(selectorLabel)
		equal := true
		for k, v := range selectorLabel {
			if podLabel[k] != v {
				equal = false
				break
			}
		}
		if equal {
			glog.Infof("resource %v %v is %v", d.GetNamespace(), d.GetName(), "daemonset")
			return true, d.GetName(), nil
		}
	}
	return false, "", nil
}

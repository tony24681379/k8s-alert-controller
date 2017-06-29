package monitoring

import (
	"net/http"
	"sync"

	"fmt"

	"github.com/golang/glog"
	"github.com/tony24681379/k8s-tools/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodRestart will scale the pod
func PodRestart(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		glog.V(2).Info("PodRestart")
		podName := r.Form.Get("pod")
		glog.V(2).Info(podName)
		namespace := r.Form.Get("namespace")
		glog.V(2).Info(namespace)
		po, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			glog.Error(err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(podName + " not found"))
			return
		}
		podLabel := po.GetLabels()
		glog.V(2).Info(podLabel)

		resourceType, resourceName, err := CheckResource(clientset, namespace, podLabel)
		if err != nil {
			glog.Error(err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(podName + " " + err.Error()))
			return
		}

		controller, err := controller.ControllerFor(clientset, resourceType, namespace)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(resourceType + " " + resourceName + " " + err.Error()))
			return
		}

		err = controller.RestartOnePod(resourceName, podName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(resourceType + " " + resourceName + " " + err.Error()))
		} else {
			w.Write([]byte(resourceType + ": " + resourceName + " namespace:" + namespace))
		}
	}
}

// CheckResource return resourceType and resourceName
func CheckResource(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (resourceType string, resourceName string, err error) {
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
	return "", "", nil
}

func checkResourceDeployment(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (bool, string, error) {
	deploymentList, err := clientset.Extensions().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return false, "", err
	}

	for _, d := range deploymentList.Items {
		deploymentLabel := d.GetLabels()
		equal := true
		for k, v := range deploymentLabel {
			if podLabel[k] != v {
				equal = false
				break
			}
		}
		glog.V(2).Info(d.GetLabels())
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
		daemonSetLabel := d.GetLabels()
		equal := true
		for k, v := range daemonSetLabel {
			if podLabel[k] != v {
				equal = false
				break
			}
		}
		glog.V(2).Info(d.GetLabels())
		if equal {
			glog.Infof("resource %v %v is %v", d.GetNamespace(), d.GetName(), "daemonset")
			return true, d.GetName(), nil
		}
	}
	return false, "", nil
}

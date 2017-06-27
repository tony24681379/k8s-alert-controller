package monitoring

import (
	"net/http"

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
		// todo err

		controller, err := controller.ControllerFor(clientset, resourceType, namespace)
		err = controller.RestartOnePod(resourceName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(resourceType + " " + resourceName + " " + err.Error()))
		} else {
			w.Write([]byte(resourceType + ": " + resourceName + " namespace:" + namespace))
		}
	}
}

// CheckResource return resourceType and resourceName
func CheckResource(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (string, string, error) {
	isDelpoyment, deploymentName, err := checkResourceDeployment(clientset, namespace, podLabel)
	if err != nil {
		return "", "", err
	}
	if isDelpoyment {
		return controller.DeploymentType, deploymentName, nil
	}
	return "", "", nil
}

func checkResourceDeployment(clientset *kubernetes.Clientset, namespace string, podLabel map[string]string) (bool, string, error) {
	deployments := clientset.Extensions().Deployments(namespace)
	deploymentList, err := deployments.List(metav1.ListOptions{})
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
			glog.Infoln(d.GetName(), d.GetNamespace(), "deployment will scale")
			return true, d.GetName(), nil
		}
	}
	return false, "", nil
}

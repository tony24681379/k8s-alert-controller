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
			glog.Error(err.Error())
		}
		podLabel := po.GetLabels()
		glog.V(2).Info(podLabel)
		deployments := clientset.Extensions().Deployments(namespace)
		deploymentList, err := deployments.List(metav1.ListOptions{})
		if err != nil {
			glog.Error(err.Error())
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
				deploymentController := &controller.DeploymentController{
					Deployment: deployments,
				}
				deploymentController.RestartOnePod(d.GetName())
				w.Write([]byte("deployment:" + d.GetName() + "namespace:" + d.GetNamespace()))
			}
		}
	}
}

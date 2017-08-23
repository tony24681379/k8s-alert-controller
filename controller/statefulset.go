package controller

import (
	"sync"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

type StatefulSetController struct {
	sync.Mutex
	Pod         v1.PodInterface
	StatefulSet v1beta1.StatefulSetInterface
}

// RestartOnePod scale up the deployment and scale down the deployment
func (d *StatefulSetController) RestartOnePod(resourceName, podName string) error {
	if err := d.deletePod(podName); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (d *StatefulSetController) deletePod(podName string) error {
	if err := d.Pod.Delete(podName, &metav1.DeleteOptions{}); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

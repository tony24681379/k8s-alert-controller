package controller

import (
	"sync"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
	types "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

type DeploymentController struct {
	sync.Mutex
	Pod        corev1types.PodInterface
	Deployment types.DeploymentInterface
}

// RestartOnePod scale up the deployment and scale down the deployment
func (d *DeploymentController) RestartOnePod(resourceName, podName string) error {
	d.Lock()
	defer d.Unlock()
	availableReplicas, err := d.updateReplicas(resourceName, 1)
	if err != nil {
		glog.Error(err)
		return err
	}

	if err := d.waitForAvailable(resourceName, availableReplicas); err != nil {
		glog.Error(err)
		return err
	}

	if err := d.deletePod(podName); err != nil {
		glog.Error(err)
	}

	if _, err := d.updateReplicas(resourceName, -1); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (d *DeploymentController) waitForAvailable(resourceName string, availableReplicas int32) error {
	glog.V(2).Info(resourceName, " ", availableReplicas)
	for {
		deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
		if err != nil {
			glog.Error(err)
			return err
		}
		time.Sleep(1 * time.Second)
		if deployment.Status.AvailableReplicas > availableReplicas {
			glog.V(2).Info(resourceName, " ", deployment.Status.AvailableReplicas)
			break
		}
	}

	return nil
}

func (d *DeploymentController) updateReplicas(resourceName string, scaleReplicas int) (int32, error) {
	deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err)
		return 0, err
	}

	*deployment.Spec.Replicas = int32(*deployment.Spec.Replicas) + int32(scaleReplicas)
	if _, err = d.Deployment.Update(deployment); err != nil {
		glog.Error(err)
		return 0, err
	}
	return deployment.Status.AvailableReplicas, nil
}

func (d *DeploymentController) deletePod(podName string) error {
	if err := d.Pod.Delete(podName, &metav1.DeleteOptions{}); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

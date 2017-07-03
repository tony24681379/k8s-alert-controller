package controller

import (
	"sync"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

type DeploymentGetter interface {
	Deployment() DeploymentController
}

type DeploymentController struct {
	sync.Mutex
	Deployment types.DeploymentInterface
}

// RestartOnePod scale up the deployment and scale down the deployment
func (d *DeploymentController) RestartOnePod(resourceName, podName string) error {
	if err := d.updateReplicas(resourceName, 1); err != nil {
		glog.Error(err)
		return err
	}

	if err := d.updateReplicas(resourceName, -1); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (d *DeploymentController) updateReplicas(resourceName string, scaleReplicas int) error {
	d.Lock()
	defer d.Unlock()
	deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err)
		return err
	}
	availableReplicas := deployment.Status.AvailableReplicas
	glog.V(2).Info(availableReplicas)
	*deployment.Spec.Replicas = int32(*deployment.Spec.Replicas) + int32(scaleReplicas)
	if _, err = d.Deployment.Update(deployment); err != nil {
		glog.Error(err)
		return err
	}

	if scaleReplicas > 0 {
		for {
			deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
			if err != nil {
				glog.Error(err)
				return err
			}
			time.Sleep(1 * time.Second)
			if deployment.Status.AvailableReplicas > availableReplicas {
				return nil
			}
		}
	}
	return nil
}

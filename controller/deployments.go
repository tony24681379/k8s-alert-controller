package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1types "k8s.io/client-go/kubernetes/typed/core/v1"
	types "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

type DeploymentController struct {
	Pod        corev1types.PodInterface
	Deployment types.DeploymentInterface
}

var deploymentMutex sync.Mutex

// RestartOnePod scale up the deployment and scale down the deployment
func (d *DeploymentController) RestartOnePod(resourceName, podName string) error {
	deploymentMutex.Lock()
	glog.V(2).Info("lock")
	defer func() {
		glog.V(2).Info("unlock")
		deploymentMutex.Unlock()
	}()

	retryTimes := 3
	for i := 0; i < retryTimes; i++ {
		err := d.updateReplicas(resourceName, 1)
		if i < retryTimes {
			if err != nil {
				glog.Error(err)
				continue
			} else {
				break
			}
		} else {
			return err
		}
	}

	if err := d.waitForAvailable(resourceName); err != nil {
		glog.Error(err)
	}

	if err := d.deletePod(podName); err != nil {
		glog.Error(err)
	}

	for i := 0; i < retryTimes; i++ {
		err := d.updateReplicas(resourceName, -1)
		if i < retryTimes {
			if err != nil {
				glog.Error(err)
				continue
			} else {
				break
			}
		} else {
			return err
		}
	}
	return nil
}

func (d *DeploymentController) waitForAvailable(resourceName string) error {
	timeOut := time.After(120 * time.Second)
	complete := make(chan bool)
	go func() {
		for {
			deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
			if err != nil {
				glog.Error(err)
				continue
			}
			time.Sleep(1 * time.Second)
			if deployment.Status.AvailableReplicas >= 1 {
				complete <- true
				break
			}
		}
	}()
	select {
	case <-timeOut:
		return fmt.Errorf("waitForAvailable time out")
	case <-complete:
		return nil
	}
}

func (d *DeploymentController) updateReplicas(resourceName string, scaleReplicas int) error {
	deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err)
		return err
	}

	*deployment.Spec.Replicas = int32(*deployment.Spec.Replicas) + int32(scaleReplicas)
	deployment, err = d.Deployment.Update(deployment)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

func (d *DeploymentController) deletePod(podName string) error {
	if err := d.Pod.Delete(podName, &metav1.DeleteOptions{}); err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

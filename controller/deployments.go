package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

type DeploymentController struct {
	Pod        v1.PodInterface
	Deployment v1beta1.DeploymentInterface
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

	deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err)
	}
	replicas := *deployment.Spec.Replicas

	if err := d.updateReplicas(resourceName, replicas+int32(1)); err != nil {
		glog.Error(err)
		return err
	}

	if err := d.waitForAvailable(resourceName); err != nil {
		glog.Error(err)
	}

	if err := d.deletePod(podName); err != nil {
		glog.Error(err)
	}

	err = d.updateReplicas(resourceName, replicas)
	if err != nil {
		glog.Error(err)
		return err
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

func (d *DeploymentController) updateReplicas(resourceName string, scaleReplicas int32) error {
	retryTimes := 3
	var err error
	for i := 0; i < retryTimes; i++ {
		deployment, err := d.Deployment.Get(resourceName, metav1.GetOptions{})
		if err != nil {
			glog.Error(err)
			time.Sleep(1 * time.Second)
			continue
		}

		*deployment.Spec.Replicas = scaleReplicas
		_, err = d.Deployment.Update(deployment)
		if err != nil {
			glog.Error(err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
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

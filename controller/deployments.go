package controller

import (
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type DeploymentController struct {
	Deployment types.DeploymentInterface
}

// RestartOnePod scale up the deployment and scale down the deployment
func (d *DeploymentController) RestartOnePod(deploymentName string) {
	deployment, err := d.Deployment.Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		glog.Error(err.Error())
	}
	availableReplicas := deployment.Status.AvailableReplicas
	d.updateReplicas(deployment, deploymentName, 1)
	deployment = d.watch(deployment, deploymentName, availableReplicas)
	d.updateReplicas(deployment, deploymentName, -1)
}

func (d *DeploymentController) updateReplicas(deployment *v1beta1.Deployment, deploymentName string, Replicas int) {
	*deployment.Spec.Replicas = int32(*deployment.Spec.Replicas) + int32(Replicas)
	_, err := d.Deployment.Update(deployment)
	if err != nil {
		glog.Error(err.Error())
	}
}

func (d *DeploymentController) watch(deployment *v1beta1.Deployment, deploymentName string, availableReplicas int32) *v1beta1.Deployment {
	glog.V(2).Info(availableReplicas)
	for {
		deployment, err := d.Deployment.Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			glog.Error(err.Error())
		}
		time.Sleep(1 * time.Second)
		if deployment.Status.AvailableReplicas > availableReplicas {
			return deployment
		}
	}
}

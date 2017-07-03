package server

import (
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/tony24681379/k8s-alert-controller/monitoring"
)

// Server serve the request
func Server(kubeconfig, port string) error {
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/pod/restart", monitoring.PodRestart(clientset))
	glog.Info("serve on port:", port)
	glog.Fatal(http.ListenAndServe(":"+port, r))
	return nil
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	glog.Info("healthz")
}

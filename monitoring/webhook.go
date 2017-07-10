package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
)

type Alert struct {
	Status string `json:"status"`
	Labels struct {
		Alertname           string `json:"alertname"`
		App                 string `json:"app"`
		Instance            string `json:"instance"`
		Job                 string `json:"job"`
		Kubernetes          string `json:"kubernetes"`
		KubernetesName      string `json:"kubernetes_name"`
		KubernetesNamespace string `json:"kubernetes_namespace"`
		Namespace           string `json:"namespace"`
		Pod                 string `json:"pod"`
		Service             string `json:"service"`
		Severity            string `json:"severity"`
	} `json:"labels"`
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
}

type Description struct {
	Service   string `json:"service"`
	Pod       string `json:"pod"`
	Namespace string `json:"namespace"`
}

type Alerts struct {
	Receiver    string  `json:"receiver"`
	Status      string  `json:"status"`
	Alerts      []Alert `json:"alerts"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	CommonLabels struct {
		Alertname  string `json:"alertname"`
		Job        string `json:"job"`
		Kubernetes string `json:"kubernetes"`
	} `json:"commonLabels"`
	CommonAnnotations struct {
	} `json:"commonAnnotations"`
	ExternalURL string `json:"externalURL"`
	Version     string `json:"version"`
	GroupKey    string `json:"groupKey"`
}

// Webhook recive prometheus alert and handle the requests
func Webhook(clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Error(err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		glog.Info(string(body))

		var alerts Alerts
		if err = json.Unmarshal(body, &alerts); err != nil {
			glog.Error("json Unmarshal fail:", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		glog.V(1).Infof("+%v", alerts)

		var res, resErr string

		for _, alert := range alerts.Alerts {
			if alert.Status == "firing" {
				go func(alert Alert) {
					r, err := handleAlert(clientset, alert)

					res = res + r
					if err != nil {
						resErr = resErr + err.Error()
					}
				}(alert)
			} else if alert.Status == "resolved" {
				continue
			}
		}
		if resErr != "" {
			w.WriteHeader(500)
			w.Write([]byte(resErr))
		} else {
			w.Write([]byte(res))
		}
	}
}

func handleAlert(clientset kubernetes.Interface, alert Alert) (result string, err error) {
	var description Description
	if err = json.Unmarshal([]byte(alert.Annotations.Description), &description); err != nil {
		glog.Error("json Unmarshal fail:", err)
		return "", err
	}
	glog.Info(alert)
	switch alert.Labels.Alertname {
	// case "PROBE_FAILED":
	case "POD_RESTART":
		result, err = PodRestart(clientset, description.Pod, description.Namespace)
	}
	return result, err
}

# k8s-alert-controller

This is a backend for prometheus alert manager.

## How to Run

```
Usage of k8s-alert-controller:
      --alsologtostderr                  log to standard error as well as files (default true)
      --kubeconfig string                absolute path of the kubeconfig file
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --port string                      serve port (default "3000")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs (default 2)
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## Run k8s-alert-controller out of kubernetes

```
$ k8s-alert-controller --kubeconfig=/.kube/config
```

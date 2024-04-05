## Introduction

All health checks are enabled by default upon installing the KubViz agent. They are automatically included, but if you don't need them, you can disable them.

```yaml
kuberhealthy:
  enabled: false
...
```

## Types of Checks

Check Name | Description |
------ | -------- | 
Daemonset check | Ensures daemonsets can be successfully deployed |
DNS status check | Checks for failures with DNS, including resolving within the cluster and outside of the cluster |
Deployment check | Ensures that a Deployment and Service can be provisioned, created, and serve traffic within the Kubernetes cluster |
Image pull check | Verifies that an image can be pulled from an image repository | 
Pod status check | Checks for unhealthy pod statuses in a target namespace |
Pod restart | Checks for excessive pod restarts in any namespace |
Resource quota check | Checks if resource quotas (CPU & memory) are available |

## Configuration

- Daemonset, Deployment, and DNS checks are enabled by default.

- Pod Status, Pod Restart, Image Pull, and Resource Quota checks need to be manually enabled.

```yaml
check:
    podRestarts:
      enabled: true
...
```

```yaml
    podStatus:
      enabled: true
...
```

```yaml
    imagePullCheck:
      enabled: true
...
```

```yaml
    resourceQuota:
      enabled: true
...
```

### Additional configuration for image-pull check 

1. Pull the test image from docker hub

```bash     
docker pull kuberhealthy/test-check
```

2. Push this image on the repository you need tested.

```bash
docker push my.repository/repo/test-check
```

- The pod is designed to attempt a pull of the test image from the remote repository (never from local). If the image is unavailable, an error will be reported to the API

### Additional configuration for resource quota check

This check tests if namespace resource quotas CPU and memory are under a specified threshold or percentage.

You need to add the namespaces to the 'WHITELIST'.

```yaml
      extraEnvs:
        BLACKLIST: "default"
        WHITELIST: "kube-system,kubviz"
...
```


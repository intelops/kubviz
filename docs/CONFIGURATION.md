# Setting Up a Webhook

In order to visualize git and container events in Kubviz, it is necessary to create a webhook for the respective repository.

You can create a webhook with your own customized data, and in the URL section, you can specify the following format.

1. The URL for a git repository will appear in the following format:

```bash
https://<INGRESS HOSTNAME>/github
```
Please replace the <INGRESS HOSTNAME> section with the specific ingress host name of your git bridge, and the path `/github` may vary depending on the git platform being used.

Possible values are:

Values | Platform |
------ | -------- | 
`/github` | GitHub |
`/gitlab` | GitLab |
`/gitea` | Gitea |
`/bitbucket` | BitBucket | 
`/azure` | Azure |

2. The URL for a Container Registry will appear in the following format:

```bash
http://<INGRESS HOSTNAME>/event/docker/hub
```

Please replace the <INGRESS HOSTNAME> section with the specific ingress host name of your container bridge, and /event/docker/hub may vary depending on the container registry platform being used.

Possible values are:

Values | Platform |
------ | -------- | 
`/event/docker/hub` | DockerHub |
`/event/azure/container` | Azure |
`/event/jfrog/container` | JFrog |
`/event/quay/container` | Quay |




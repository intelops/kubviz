# Creating Webhook in GitLab

Creating a webhook in GitLab will fetch all the git events from your repository, and this data will be passed to Kubviz.

Follow the below steps to create a webhook in GitLab

1. Select you repository which you want to create a webhook.

2. Open the repository and navigate to the Settings option. From there, select the Webhooks option.

Settings ---> Webhooks

3. Inside URL area give the gitbridge ingress host name

```bash
https://gitbridge.example/gitlab
```

4. Leave the secret token area as blank.

5. Select which events would you like to trigger this webhook?

6. Untick the SSL verification check box click Add webhook button it will create a webhook for your repository.
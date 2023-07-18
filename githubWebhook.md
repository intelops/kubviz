# Creating Webhook in GitHub

Creating a webhook in GitHub will fetch all the git events from your repository, and this data will be passed to Kubviz.

Follow the below steps to create a webhook in GitHub

1. Select you repository which you want to create a webhook.

2. Open the repository, navigate to the Settings option, and then select the Webhooks option

Settings ---> Webhooks --> Add webhooks.

3. Inside Payload URL area, give the gitbridge ingress host name:

```bash
https://gitbridge.example/github
```

4. Select the content type as application/json.

5. Leave the secret area as blank.

6. Select which events would you like to trigger this webhook?

7. Click the Add webhook button it will create a webhook for your repository.
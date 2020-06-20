# godoc-service

This service publishes godoc for private github repositories.

To install in your own cluser:

1. Fork this repo

2. Create a secret in your cluster named `godoc-service-config` with the following keys:

```
apiVersion: v1
kind: Secret
metadata:
  name: godoc-service-config
  namespace: default
type: Opaque
data:
  githubToken: <github token with read access to your repos>
  githubRepos: <org1>/<repo1>;<org2>/<repo2> ...
```

3. Follow the instructions [here](https://www.getambassador.io/docs/latest/tutorials/projects/) to
   Create an Ambassador Project Resource that refers to your forked copy of the repo.

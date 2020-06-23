# Godoc Service

This service publishes [godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc?tab=doc) for private GitHub repositories and designed to be deployed using Kubernetes and [Ambassador Edge Stack](https://www.getambassador.io).

## Installation and deployment instructions

1. Get a GitHub personal access token with access to your repository.
2. Create a secret in your cluster named `godoc-service-config` with the token and the repositories you want to access:


```
kubectl create secret generic godoc-service-config --from-literal=githubToken=a17531d296845a1c16dd67df38065c1ee55c067 --from-literal=githubRepos="ambassadorlabs/godoc-service;datawire/ambassador"
```

3. Create an Ambassador [Project resource](https://www.getambassador.io/docs/latest/tutorials/projects/) to build and deploy this service directly into Kubernetes from GitHub:

```
---
apiVersion: getambassador.io/v2
kind: Project
metadata:
 name: my-doc-repo
 namespace: default
spec:
 host:  objective-nash-360.edgestack.me
 prefix: /doc/
 githubRepo: ambassadorlabs/godoc-service
 githubToken:  a17531d296845a1c16dd67df38065c1ee55c067
```

4. Create an Ambassador [Filter](https://www.getambassador.io/docs/latest/topics/using/filters/) to enable Single Sign-On.
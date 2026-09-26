# Knowledge Hub Helm chart

The chart deploys the API, frontend, document reader, PostgreSQL, Redis, and optional supporting services. Review [values.yaml](values.yaml) before installation, especially image repositories, credentials, ingress host, storage, and resource limits.

```sh
helm lint ./helm
helm install knowledge-hub ./helm -f your-values.yaml
```

The default image repositories are `knowledge-hub-app`, `knowledge-hub-ui`, and `knowledge-hub-docreader`. Push these images to a registry you control and set the corresponding repository values before deploying to a cluster.

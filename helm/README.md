# Knowledge Hub Helm chart

The chart deploys the API, frontend, document reader, PostgreSQL, Redis, and optional supporting services. Review [values.yaml](values.yaml) before installation, especially image repositories, credentials, ingress host, storage, and resource limits.

```sh
helm lint ./helm
helm install knowledge-hub ./helm -f your-values.yaml
```

The default image repositories are `knowledge-hub-app`, `knowledge-hub-ui`, and `knowledge-hub-docreader`. Push these images to a registry you control and set the corresponding repository values before deploying to a cluster.

## Service names

Optional `<component>.service.name` values override the Service names of `app`, `frontend`, `postgresql`, `redis`, `docreader`, and `neo4j`. Omitted or empty values keep the defaults (`app`, `frontend`, `postgres`, `redis`, `docreader`, `neo4j`). Backend addresses, Ingress, and installation notes follow the overrides; the frontend `APP_HOST` follows the App Service unless `frontend.appHost` is set.

Use distinct names for every enabled Service when you install several releases in one namespace, because the defaults are not release-prefixed. Renaming a Service on an existing installation requires a DNS and consumer cutover plan.

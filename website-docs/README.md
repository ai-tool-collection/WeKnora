# Knowledge Hub documentation site

The documentation site uses VitePress and publishes under `/docs/`. Its root page redirects there. The documentation is in English.

```sh
cd website-docs
npm ci
npm run build
npm run preview
```

`npm run build` checks local Markdown links, builds the VitePress pages, and writes `static-site/`. Use `npm run package:site` to create a deployment archive. The Dockerfile builds and serves the same directory with nginx.

> ARCHIVED - This code has moved to pentops/j5

Registry
========

A central registry for API versions.


Server
------

### Go Mod Proxy

`/gopkg/`

Implements a Go Module Proxy.

Currently serves pre-cached go modules only, A later version will allow pull-through cache.

### JSON API Registry

`/registry/$org/$image/$version/$format`

Where format is one of:

- `swagger.json`
- `jdef.json`
- `api.json`
- `image.bin`



### Local Proto Go Builder

Based on a config file very similar to buf.gen.yaml, generates go modules from
source, and adds them to the proxy store.

Currently this works only as a local command, however this will be linked to a github action / remote builder.

![pipeline](./ext/images/pipeline.svg)

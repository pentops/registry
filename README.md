Registry
========

A central registry for API versions.

- A local builder for protoc-gen- output into go modules


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
- `image.bin`



### Local Proto Go Builder

Based on a config file very similar to buf.gen.yaml, generates go modules from
source, and adds them to the proxy store.



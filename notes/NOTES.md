# Notes

## Kustomize

*nameReference* - Give a source resource, replace its name 
  into several other resources at the given path.

*namespace* - Give a source resource, replace its namespace 
  into several other resources at the given path.

*varReference* - Replace variable reference in the format `$(FOO)`
  with a value from a field in some other resource.

## Removed / TODO

- RBAC proxy
- Prometheus CRDs
- certmanager
- webhook

CSV generator does not support unofficial CSV fields such as `spec.relatedImages`.
See https://github.com/operator-framework/operator-sdk/issues/4503 . 
**TODO** Figure out how to include it.

Can not rename `config/manifests/bases` to `config/manifests/resources`.

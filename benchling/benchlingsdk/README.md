# benchlingsdk

This package provides a client SDK binding to the benchling API that is
generated from benchling's openapi spec as shown below. Unfortunately
the spec contains various bugs and irregularities which need to be fixed
before the ```oapi-codegen``` generator can use it. ```oapi-tool``` is used to
perform these transformations according to the ```transformations.yaml``` configuration file.

```sh
go install -x github.com/cosnicolaou/oapi-tool@latest
datestamp=$(date +'%m-%d-%Y')
spec="benchling-${datestamp}.yaml"
formatted="benchling-formatted-${datestamp}.yaml"
transformed="benchling-transformed-${datestamp}.yaml"
oapi-tool download --output="${spec}" 'https://benchling.com/api/v2/openapi.yaml'
oapi-tool format --output="${formatted}" -validate=false "${spec}"
oapi-tool transform --output="${transformed}" --config=transformations.yaml ${formatted}
# this is the earliest usable version, v1.12.5 onward should be fine.
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@f4cf8f9
oapi-codegen --package=benchlingsdk "${transformed}" > benchlingsdk.go
```

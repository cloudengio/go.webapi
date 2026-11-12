# benchlingsdk

This package provides a client SDK binding to the benchling API that is
generated from benchling's openapi spec as shown below. Unfortunately
the spec contains various bugs and irregularities which need to be fixed
before the ```oapi-codegen``` generator can use it. ```oapi-tool``` is used to
perform these transformations according to the ```transformations.yaml``` configuration file.

```sh
go install cloudeng.io/webapi/oapi-tool@latest
datestamp=$(date +'%m-%d-%Y')
spec="benchling-${datestamp}.yaml"
formatted="benchling-formatted-${datestamp}.yaml"
transformed="benchling-transformed-${datestamp}.yaml"
oapi-tool download --output="${spec}" 'https://benchling.com/api/v2/openapi.yaml'
oapi-tool format --output="${formatted}" -validate=false "${spec}"
oapi-tool transform --output="${transformed}" --config=benchling-transformations.yaml "${formatted}"
oapi-tool validate "${transformed}"

# oapi-codegen v2 supports echo v5 via the echo5-server generate option
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen -generate "types,client,echo5-server" --package=benchlingsdk "${transformed}" > benchlingsdk.go
```

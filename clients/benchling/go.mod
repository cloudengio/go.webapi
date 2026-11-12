module cloudeng.io/webapi/clients/benchling

go 1.27.0

require (
	cloudeng.io/algo v0.0.0-20261111225710-e0ef01d9df9b
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20261111225710-e0ef01d9df9b
	cloudeng.io/logging v0.0.0-20261111225710-e0ef01d9df9b
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/sync v0.0.12-0.20260804222138-e9281ed260ba
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	github.com/getkin/kin-openapi v0.153.0
	github.com/labstack/echo/v4 v4.16.3
	github.com/oapi-codegen/runtime v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/cmdutil v0.0.0-20261111225710-e0ef01d9df9b // indirect
	cloudeng.io/os v0.0.0-20261111225710-e0ef01d9df9b // indirect
	cloudeng.io/text v0.0.16-0.20260624171915-da98fe9dec2b // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	go.yaml.in/yaml/v4 v4.0.0 // indirect
	golang.org/x/crypto v0.59.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/oauth2 v0.38.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

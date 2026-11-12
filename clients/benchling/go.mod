module cloudeng.io/webapi/clients/benchling

go 1.27.0

require (
	cloudeng.io/algo v0.0.0-20261112002641-3f0b5fb85a24
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20261112002641-3f0b5fb85a24
	cloudeng.io/logging v0.0.0-20261112002641-3f0b5fb85a24
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/sync v0.0.12-0.20260804222138-e9281ed260ba
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	github.com/labstack/echo/v5 v5.4.0
	github.com/oapi-codegen/runtime v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/cmdutil v0.0.0-20261112002641-3f0b5fb85a24 // indirect
	cloudeng.io/os v0.0.0-20261112002641-3f0b5fb85a24 // indirect
	cloudeng.io/text v0.0.16-0.20260624171915-da98fe9dec2b // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/oauth2 v0.38.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/text v0.44.0 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

module cloudeng.io/webapi/oapi-tool

go 1.26

require (
	cloudeng.io/cmdutil v0.0.0-20260605174237-2d6c1041426f
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/openapi v0.0.0-20260512044422-94ea35672b76
	github.com/getkin/kin-openapi v0.140.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/file v0.0.0-20260605174237-2d6c1041426f // indirect
	cloudeng.io/text v0.0.16-0.20260312171538-61fcde6ce278 // indirect
	github.com/go-openapi/jsonpointer v0.23.1 // indirect
	github.com/go-openapi/swag/jsonname v0.26.0 // indirect
	github.com/oasdiff/yaml v0.1.0 // indirect
	github.com/oasdiff/yaml3 v0.0.13 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/text v0.37.0 // indirect
)

replace cloudeng.io/webapi/openapi => ../openapi

module cloudeng.io/webapi/oapi-tool

go 1.26.4

require (
	cloudeng.io/cmdutil v0.0.0-20260721222700-155e56185eeb
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/openapi v0.0.0-20260512044422-94ea35672b76
	github.com/getkin/kin-openapi v0.143.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/algo v0.0.0-20260721222700-155e56185eeb // indirect
	cloudeng.io/file v0.0.0-20260721222700-155e56185eeb // indirect
	cloudeng.io/sync v0.0.11 // indirect
	cloudeng.io/sys v0.0.0-20260721222700-155e56185eeb // indirect
	cloudeng.io/text v0.0.16-0.20260624171915-da98fe9dec2b // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace cloudeng.io/webapi/openapi => ../openapi

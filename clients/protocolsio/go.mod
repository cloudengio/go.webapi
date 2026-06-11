module cloudeng.io/webapi/clients/protocolsio

go 1.26

require (
	cloudeng.io/cmdutil v0.0.0-20260611161950-23029f4a5674
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20260611161950-23029f4a5674
	cloudeng.io/logging v0.0.0-20260611161950-23029f4a5674
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/algo v0.0.0-20260611161950-23029f4a5674 // indirect
	cloudeng.io/os v0.0.0-20260611161950-23029f4a5674 // indirect
	cloudeng.io/sync v0.0.11 // indirect
	cloudeng.io/text v0.0.16-0.20260312171538-61fcde6ce278 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

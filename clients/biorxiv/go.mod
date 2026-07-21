module cloudeng.io/webapi/clients/biorxiv

go 1.26.4

require (
	cloudeng.io/algo v0.0.0-20260721222700-155e56185eeb
	cloudeng.io/cmdutil v0.0.0-20260721222700-155e56185eeb
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20260721222700-155e56185eeb
	cloudeng.io/logging v0.0.0-20260721222700-155e56185eeb
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/os v0.0.0-20260721222700-155e56185eeb // indirect
	cloudeng.io/sync v0.0.11 // indirect
	cloudeng.io/sys v0.0.0-20260721222700-155e56185eeb // indirect
	cloudeng.io/text v0.0.16-0.20260624171915-da98fe9dec2b // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

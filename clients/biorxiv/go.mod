module cloudeng.io/webapi/clients/biorxiv

go 1.26

require (
	cloudeng.io/algo v0.0.0-20260601212956-cbb9ba55a911
	cloudeng.io/cmdutil v0.0.0-20260601212956-cbb9ba55a911
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20260601212956-cbb9ba55a911
	cloudeng.io/logging v0.0.0-20260601212956-cbb9ba55a911
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
)

require (
	cloudeng.io/os v0.0.0-20260601212956-cbb9ba55a911 // indirect
	cloudeng.io/webapi/webapitestutil v0.0.0-20260108223722-702b7fae5336 // indirect
)

require (
	cloudeng.io/sync v0.0.11 // indirect
	cloudeng.io/text v0.0.16-0.20260312171538-61fcde6ce278 // indirect
	golang.org/x/net v0.55.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace cloudeng.io/webapi/operations => ../../operations

module cloudeng.io/webapi/clients/protocolsio

go 1.27.0

require (
	cloudeng.io/cmdutil v0.0.0-20261018232534-249a4dd7d02a
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278
	cloudeng.io/file v0.0.0-20261018232534-249a4dd7d02a
	cloudeng.io/logging v0.0.0-20261018232534-249a4dd7d02a
	cloudeng.io/path v0.0.10-0.20260312171538-61fcde6ce278
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloudeng.io/algo v0.0.0-20261018232534-249a4dd7d02a // indirect
	cloudeng.io/os v0.0.0-20261018232534-249a4dd7d02a // indirect
	cloudeng.io/sync v0.0.12-0.20260804222138-e9281ed260ba // indirect
	cloudeng.io/sys v0.0.0-20260925200243-4023f51f1ebd // indirect
	cloudeng.io/text v0.0.16-0.20260624171915-da98fe9dec2b // indirect
	cloudeng.io/types v0.0.0-20261018232534-249a4dd7d02a // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

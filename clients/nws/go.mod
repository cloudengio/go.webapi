module cloudeng.io/webapi/clients/nws

go 1.26

require (
	cloudeng.io/algo v0.0.0-20260513030906-195852faa5d0
	cloudeng.io/datetime v0.0.0-20260513030906-195852faa5d0
	cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a
	cloudeng.io/webapi/webapitestutil v0.0.0-20260512044422-94ea35672b76
)

require (
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278 // indirect
	cloudeng.io/file v0.0.0-20260513030906-195852faa5d0 // indirect
	cloudeng.io/logging v0.0.0-20260513030906-195852faa5d0 // indirect
	cloudeng.io/sync v0.0.10 // indirect
	cloudeng.io/sys v0.0.0-20260511051612-3406526ae677 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

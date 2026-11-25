module cloudeng.io/webapi/clients/openai

go 1.27.0

require cloudeng.io/webapi/operations v0.0.0-20260510204434-243224b8f05a

require (
	cloudeng.io/algo v0.0.0-20261112002641-3f0b5fb85a24 // indirect
	cloudeng.io/errors v0.0.14-0.20260312171538-61fcde6ce278 // indirect
	cloudeng.io/file v0.0.0-20261118164247-00ff6e0d0bfd // indirect
	cloudeng.io/logging v0.0.0-20261112002641-3f0b5fb85a24 // indirect
	cloudeng.io/sync v0.0.12-0.20261112002641-3f0b5fb85a24 // indirect
)

replace cloudeng.io/webapi/operations => ../../operations

replace cloudeng.io/webapi/webapitestutil => ../../webapitestutil

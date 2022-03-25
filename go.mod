module main

go 1.17

replace a10/axapi => ./axapi

require (
	a10/axapi v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/tidwall/gjson v1.11.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
)

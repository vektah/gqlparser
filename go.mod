module github.com/vektah/gqlparser/v2

go 1.22

require (
	github.com/agnivade/levenshtein v1.2.1
	github.com/stretchr/testify v1.12.0
	go.yaml.in/yaml/v3 v3.0.5
)

require (
	github.com/kr/pretty v0.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v2.5.14
	v2.5.13
)

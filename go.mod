module github.com/vektah/gqlparser/v2

go 1.22

require (
	github.com/agnivade/levenshtein v1.2.1
	github.com/stretchr/testify v1.11.1
	go.yaml.in/yaml/v3 v3.0.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v2.5.14
	v2.5.13
)

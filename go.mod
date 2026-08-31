module github.com/vektah/gqlparser/v2

go 1.22

require (
	github.com/agnivade/levenshtein v1.2.1
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v3 v3.0.5
)

retract (
	v2.5.14
	v2.5.13
)

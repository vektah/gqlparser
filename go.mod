module github.com/vektah/gqlparser/v2

go 1.23

require (
	github.com/99designs/gqlgen v0.17.64
	github.com/agnivade/levenshtein v1.2.1
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
)

retract (
	v2.5.14
	v2.5.13
)

package testrunner

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/vektah/gqlparser/v2/validator"
)

// this registry is used to build
var testErrorRegistry = []error{
	new(validator.EmptyDefinitionError),
}

type registeredError struct {
	ErrType reflect.Type

	// RegistryIndex is the index in [testErrorRegistry]
	// from which this [registeredError] originates.
	RegistryIndex int
}

var errorRegistry = make(map[string]registeredError)

func init() {
	for i, err := range testErrorRegistry {
		original := reflect.TypeOf(err)
		t := original
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		pkgPath := t.PkgPath()
		typeName := t.Name()
		qualifiedName := fmt.Sprintf("%s.%s", pkgPath, typeName)
		if typeName == "" {
			panic(fmt.Sprintf("unsupported error in testErrorRegistry at index %d: only named types are supported", i))
		}

		if prev, ok := errorRegistry[qualifiedName]; ok {
			panic(fmt.Sprintf("duplicate error definitions in testErrorRegistry at indices%d and %d", prev.RegistryIndex, i))
		}

		errorRegistry[qualifiedName] = registeredError{
			ErrType:       original,
			RegistryIndex: i,
		}
	}
}

func assertErrorAs(t *testing.T, err error, asQualifiedTypes []string) {
	t.Helper()

	for i, qualifiedErr := range asQualifiedTypes {
		registeredError, ok := errorRegistry[qualifiedErr]
		if !ok {
			t.Errorf(
				"unsupported typename %q at index %d: has the type been added to the testErrorRegistry?",
				qualifiedErr,
				i,
			)
			continue
		}

		asErr := reflect.New(registeredError.ErrType)
		if !errors.As(err, asErr.Interface()) {
			t.Errorf(
				"expected error to be or unwrap into %q (found at index %d): got %T",
				asErr.Type().String(),
				i,
				err,
			)
		}
	}
}

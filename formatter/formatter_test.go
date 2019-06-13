package formatter_test

import (
	"bytes"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
	"github.com/vektah/gqlparser/parser"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFormatter_FormatSchema(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchema"

	fs, err := ioutil.ReadDir(testSourceDir)
	if err != nil {
		t.Fatal(fs)
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		f := f

		t.Run(f.Name(), func(t *testing.T) {
			// load stuff
			schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
				Name:  f.Name(),
				Input: mustReadFile(path.Join(testSourceDir, f.Name())),
			})
			if gqlErr != nil {
				t.Fatal(gqlErr)
			}

			// exec format
			var buf bytes.Buffer
			err = formatter.NewFormatter(&buf).FormatSchema(schema)
			if err != nil {
				t.Fatal(err)
			}

			// validity check
			_, gqlErr = gqlparser.LoadSchema(&ast.Source{
				Name:  f.Name(),
				Input: buf.String(),
			})
			if gqlErr != nil {
				t.Log(buf.String())
				t.Fatal(gqlErr)
			}

			// golden testing
			expectedFilePath := path.Join(testBaselineDir, f.Name())
			expected, err := ioutil.ReadFile(expectedFilePath)
			if os.IsNotExist(err) {
				err = os.MkdirAll(testBaselineDir, 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = ioutil.WriteFile(expectedFilePath, buf.Bytes(), 0444)
				if err != nil {
					t.Fatal(err)
				}
				return
			} else if err != nil {
				t.Fatal(err)
			}

			if string(expected) == buf.String() {
				return
			}

			diff := difflib.UnifiedDiff{
				A:       difflib.SplitLines(string(expected)),
				B:       difflib.SplitLines(buf.String()),
				Context: 5,
			}
			d, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("if you want to accept new result. rm -rf testdata/baseline")
			t.Error(d)
		})
	}
}

func TestFormatter_FormatSchemaDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchemaDocument"

	fs, err := ioutil.ReadDir(testSourceDir)
	if err != nil {
		t.Fatal(fs)
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		f := f

		t.Run(f.Name(), func(t *testing.T) {
			// load stuff
			doc, gqlErr := parser.ParseSchema(&ast.Source{
				Name:  f.Name(),
				Input: mustReadFile(path.Join(testSourceDir, f.Name())),
			})
			if gqlErr != nil {
				t.Fatal(gqlErr)
			}

			// exec format
			var buf bytes.Buffer
			err = formatter.NewFormatter(&buf).FormatSchemaDocument(doc)
			if err != nil {
				t.Fatal(err)
			}

			// validity check
			_, gqlErr = parser.ParseSchema(&ast.Source{
				Name:  f.Name(),
				Input: buf.String(),
			})
			if gqlErr != nil {
				t.Log(buf.String())
				t.Fatal(gqlErr)
			}

			// golden testing
			expectedFilePath := path.Join(testBaselineDir, f.Name())
			expected, err := ioutil.ReadFile(expectedFilePath)
			if os.IsNotExist(err) {
				err = os.MkdirAll(testBaselineDir, 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = ioutil.WriteFile(expectedFilePath, buf.Bytes(), 0444)
				if err != nil {
					t.Fatal(err)
				}
				return
			} else if err != nil {
				t.Fatal(err)
			}

			if string(expected) == buf.String() {
				return
			}

			diff := difflib.UnifiedDiff{
				A:       difflib.SplitLines(string(expected)),
				B:       difflib.SplitLines(buf.String()),
				Context: 5,
			}
			d, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("if you want to accept new result. rm -rf testdata/baseline")
			t.Error(d)
		})
	}
}

func TestFormatter_FormatQueryDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/query"
	const testBaselineDir = "./testdata/baseline/FormatQueryDocument"

	fs, err := ioutil.ReadDir(testSourceDir)
	if err != nil {
		t.Fatal(fs)
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		f := f

		t.Run(f.Name(), func(t *testing.T) {
			// load stuff
			doc, gqlErr := parser.ParseQuery(&ast.Source{
				Name:  f.Name(),
				Input: mustReadFile(path.Join(testSourceDir, f.Name())),
			})
			if gqlErr != nil {
				t.Fatal(gqlErr)
			}

			// exec format
			var buf bytes.Buffer
			err = formatter.NewFormatter(&buf).FormatQueryDocument(doc)
			if err != nil {
				t.Fatal(err)
			}

			// validity check
			_, gqlErr = parser.ParseQuery(&ast.Source{
				Name:  f.Name(),
				Input: buf.String(),
			})
			if gqlErr != nil {
				t.Log(buf.String())
				t.Fatal(gqlErr)
			}

			// golden testing
			expectedFilePath := path.Join(testBaselineDir, f.Name())
			expected, err := ioutil.ReadFile(expectedFilePath)
			if os.IsNotExist(err) {
				err = os.MkdirAll(testBaselineDir, 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = ioutil.WriteFile(expectedFilePath, buf.Bytes(), 0444)
				if err != nil {
					t.Fatal(err)
				}
				return
			} else if err != nil {
				t.Fatal(err)
			}

			if string(expected) == buf.String() {
				return
			}

			diff := difflib.UnifiedDiff{
				A:       difflib.SplitLines(string(expected)),
				B:       difflib.SplitLines(buf.String()),
				Context: 5,
			}
			d, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("if you want to accept new result. rm -rf testdata/baseline")
			t.Error(d)
		})
	}
}

func mustReadFile(name string) string {
	src, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return string(src)
}

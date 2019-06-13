package formatter_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"unicode/utf8"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
	"github.com/vektah/gqlparser/parser"
)

func TestFormatter_FormatSchema(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchema"

	executeGoldenTesting(t, &goldenConfig{
		SourceDir: testSourceDir,
		BaselineFileName: func(cfg *goldenConfig, f os.FileInfo) string {
			return path.Join(testBaselineDir, f.Name())
		},
		Run: func(t *testing.T, cfg *goldenConfig, f os.FileInfo) []byte {
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
			err := formatter.NewFormatter(&buf).FormatSchema(schema)
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

			return buf.Bytes()
		},
	})
}

func TestFormatter_FormatSchemaDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchemaDocument"

	executeGoldenTesting(t, &goldenConfig{
		SourceDir: testSourceDir,
		BaselineFileName: func(cfg *goldenConfig, f os.FileInfo) string {
			return path.Join(testBaselineDir, f.Name())
		},
		Run: func(t *testing.T, cfg *goldenConfig, f os.FileInfo) []byte {
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
			err := formatter.NewFormatter(&buf).FormatSchemaDocument(doc)
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

			return buf.Bytes()
		},
	})
}

func TestFormatter_FormatQueryDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/query"
	const testBaselineDir = "./testdata/baseline/FormatQueryDocument"

	executeGoldenTesting(t, &goldenConfig{
		SourceDir: testSourceDir,
		BaselineFileName: func(cfg *goldenConfig, f os.FileInfo) string {
			return path.Join(testBaselineDir, f.Name())
		},
		Run: func(t *testing.T, cfg *goldenConfig, f os.FileInfo) []byte {
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
			err := formatter.NewFormatter(&buf).FormatQueryDocument(doc)
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

			return buf.Bytes()
		},
	})
}

type goldenConfig struct {
	SourceDir        string
	IsTarget         func(f os.FileInfo) bool
	BaselineFileName func(cfg *goldenConfig, f os.FileInfo) string
	Run              func(t *testing.T, cfg *goldenConfig, f os.FileInfo) []byte
}

func executeGoldenTesting(t *testing.T, cfg *goldenConfig) {
	t.Helper()

	if cfg.IsTarget == nil {
		cfg.IsTarget = func(f os.FileInfo) bool {
			return !f.IsDir()
		}
	}
	if cfg.BaselineFileName == nil {
		t.Fatal("BaselineFileName function is required")
	}
	if cfg.Run == nil {
		t.Fatal("Run function is required")
	}

	fs, err := ioutil.ReadDir(cfg.SourceDir)
	if err != nil {
		t.Fatal(fs)
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		f := f

		t.Run(f.Name(), func(t *testing.T) {
			result := cfg.Run(t, cfg, f)

			expectedFilePath := cfg.BaselineFileName(cfg, f)
			expected, err := ioutil.ReadFile(expectedFilePath)
			if os.IsNotExist(err) {
				err = os.MkdirAll(path.Dir(expectedFilePath), 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = ioutil.WriteFile(expectedFilePath, result, 0444)
				if err != nil {
					t.Fatal(err)
				}
				return
			} else if err != nil {
				t.Fatal(err)
			}

			if bytes.Equal(expected, result) {
				return
			}

			t.Logf("if you want to accept new result. rm %s", expectedFilePath)

			if utf8.Valid(expected) {
				diff := difflib.UnifiedDiff{
					A:       difflib.SplitLines(string(expected)),
					B:       difflib.SplitLines(string(result)),
					Context: 5,
				}
				d, err := difflib.GetUnifiedDiffString(diff)
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
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

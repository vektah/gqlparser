package formatter_test

import (
	"bytes"
	"flag"
	"os"
	"path"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
)

var update = flag.Bool("u", false, "update golden files")

var optionSets = []struct {
	name string
	opts []formatter.FormatterOption
}{
	{"default", nil},
	{"spaceIndent", []formatter.FormatterOption{formatter.WithIndent(" ")}},
	{"comments", []formatter.FormatterOption{formatter.WithComments()}},
	{"no_description", []formatter.FormatterOption{formatter.WithoutDescription()}},
	{"compacted", []formatter.FormatterOption{formatter.WithCompacted()}},
}

func TestFormatter_FormatSchema(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchema"

	for _, optionSet := range optionSets {
		testBaselineDir := filepath.Join(testBaselineDir, optionSet.name)
		opts := optionSet.opts
		t.Run(optionSet.name, func(t *testing.T) {
			executeGoldenTesting(t, &goldenConfig{
				SourceDir: testSourceDir,
				BaselineFileName: func(cfg *goldenConfig, f os.DirEntry) string {
					return path.Join(testBaselineDir, f.Name())
				},
				Run: func(t *testing.T, cfg *goldenConfig, f os.DirEntry) []byte {
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
					formatter.NewFormatter(&buf, opts...).FormatSchema(schema)

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
		})
	}
}

func TestFormatter_FormatSchemaDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/schema"
	const testBaselineDir = "./testdata/baseline/FormatSchemaDocument"

	for _, optionSet := range optionSets {
		testBaselineDir := filepath.Join(testBaselineDir, optionSet.name)
		opts := optionSet.opts
		t.Run(optionSet.name, func(t *testing.T) {
			executeGoldenTesting(t, &goldenConfig{
				SourceDir: testSourceDir,
				BaselineFileName: func(cfg *goldenConfig, f os.DirEntry) string {
					return path.Join(testBaselineDir, f.Name())
				},
				Run: func(t *testing.T, cfg *goldenConfig, f os.DirEntry) []byte {
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
					formatter.NewFormatter(&buf, opts...).FormatSchemaDocument(doc)

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
		})
	}
}

func TestFormatter_FormatQueryDocument(t *testing.T) {
	const testSourceDir = "./testdata/source/query"
	const testBaselineDir = "./testdata/baseline/FormatQueryDocument"

	for _, optionSet := range optionSets {
		testBaselineDir := filepath.Join(testBaselineDir, optionSet.name)
		opts := optionSet.opts
		t.Run(optionSet.name, func(t *testing.T) {
			executeGoldenTesting(t, &goldenConfig{
				SourceDir: testSourceDir,
				BaselineFileName: func(cfg *goldenConfig, f os.DirEntry) string {
					return path.Join(testBaselineDir, f.Name())
				},
				Run: func(t *testing.T, cfg *goldenConfig, f os.DirEntry) []byte {
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
					formatter.NewFormatter(&buf, opts...).FormatQueryDocument(doc)

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
		})
	}
}

type goldenConfig struct {
	SourceDir        string
	IsTarget         func(f os.FileInfo) bool
	BaselineFileName func(cfg *goldenConfig, f os.DirEntry) string
	Run              func(t *testing.T, cfg *goldenConfig, f os.DirEntry) []byte
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

	fs, err := os.ReadDir(cfg.SourceDir)
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

			if *update {
				err := os.Remove(expectedFilePath)
				if err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}

			expected, err := os.ReadFile(expectedFilePath)
			if os.IsNotExist(err) {
				err = os.MkdirAll(path.Dir(expectedFilePath), 0o755)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(expectedFilePath, result, 0o444)
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

			if utf8.Valid(expected) {
				assert.Equalf(t, string(expected), string(result), "if you want to accept new result. use -u option")
			}
		})
	}
}

func mustReadFile(name string) string {
	src, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return string(src)
}

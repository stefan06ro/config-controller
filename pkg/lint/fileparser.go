package lint

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"text/template/parse"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"

	"github.com/giantswarm/microerror"
	pathmodifier "github.com/giantswarm/valuemodifier/path"
)

var (
	fMap                 = dummyFuncMap()
	includes             = &includeExtract{[]string{}}
	templatePathPattern  = regexp.MustCompile(`(\.[a-zA-Z].[a-zA-Z0-9_\.]+)`)
	yamlErrorLinePattern = regexp.MustCompile(`yaml: line (\d+)`)
)

func init() {
	fMap["include"] = includes.include
}

type ValueFile struct {
	filepath     string
	installation string // optional
	paths        map[string]*ValuePath
	sourceBytes  []byte
}

type ValuePath struct {
	Value interface{}
	// files using this value
	UsedBy []*TemplateFile
	// value is overshadowed by some files
	OvershadowedBy []*ValueFile
}

type TemplateFile struct {
	filepath     string
	installation string // optional for defaults
	app          string

	// values map contains values requested in template using template's dot
	// notation, e.g. '{{ .some.value }}'
	values map[string]*TemplateValue
	// paths map contains all paths in template extracted by valuemodifier/path
	paths map[string]bool
	// includes contains names of all include files used by this template
	includes []string

	sourceBytes    []byte
	sourceTemplate *template.Template
}

type TemplateValue struct {
	Path            string
	OccurrenceCount int
	// MayBeMissing is set when value is not found in config.
	// Linter will check if it's patched in by any of the template patches. If
	// yes, fine. If not, that's an error and linter will let you know.
	MayBeMissing bool
}

func NewValueFile(filepath string, body []byte) (*ValueFile, error) {
	if !strings.HasSuffix(filepath, ".yaml") && !strings.HasSuffix(filepath, ".yaml.patch") {
		return nil, microerror.Maskf(executionFailedError, "given file is not a value file: %q", filepath)
	}

	// extract paths with valuemodifier path service
	allPaths := map[string]*ValuePath{}
	{
		c := pathmodifier.Config{
			InputBytes: body,
			Separator:  ".",
		}
		svc, err := pathmodifier.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		paths, err := svc.All()
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "error getting all paths for %q", filepath)
		}

		for _, path := range paths {
			value, err := svc.Get(path)
			if err != nil {
				return nil, microerror.Maskf(executionFailedError, "error getting %q value for %q: %s", filepath, path, err)
			}

			v := ValuePath{
				Value:          value,
				UsedBy:         []*TemplateFile{},
				OvershadowedBy: []*ValueFile{},
			}
			allPaths[NormalPath(path)] = &v
		}
	}

	vf := &ValueFile{
		filepath:    filepath,
		paths:       allPaths,
		sourceBytes: body,
	}

	// assign installation if possible
	if strings.HasPrefix(filepath, "installations") {
		elements := strings.Split(filepath, "/")
		vf.installation = elements[1]
	}

	return vf, nil
}

func NewTemplateFile(filepath string, body []byte) (*TemplateFile, error) {
	if !strings.HasSuffix(filepath, ".template") && !strings.HasSuffix(filepath, "values.yaml.patch") {
		return nil, microerror.Maskf(executionFailedError, "given file is not a template: %q", filepath)
	}

	tf := &TemplateFile{
		filepath:    filepath,
		sourceBytes: body,
	}

	// extract templated values and all paths from the template
	values := map[string]*TemplateValue{}
	paths := map[string]bool{}
	{
		includes.clear()
		t, err := template.
			New(filepath).
			Funcs(fMap).
			Option("missingkey=zero").
			Parse(string(body))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tf.sourceTemplate = t

		// extract all values
		for _, node := range t.Tree.Root.Nodes {
			if node.Type() == parse.NodeText {
				continue
			}

			nodePaths := templatePathPattern.FindAllString(node.String(), -1)
			for _, np := range nodePaths {
				normalPath := NormalPath(np)
				if _, ok := values[normalPath]; !ok {
					values[normalPath] = &TemplateValue{
						Path:            normalPath,
						OccurrenceCount: 1,
					}
				} else {
					values[normalPath].OccurrenceCount += 1
				}
			}
		}

		// extract all paths
		output := bytes.NewBuffer([]byte{})
		var data interface{}
		// Render template without values. All templated values will be
		// replaced by default zero values: "" for string, 0 for int, false
		// for bool etc.
		err = t.Execute(output, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tf.includes = includes.Filepaths

		c := pathmodifier.Config{
			InputBytes: output.Bytes(),
			Separator:  ".",
		}

		svc, err := pathmodifier.New(c)
		if err != nil {
			// try to pretty print offending yaml
			var yamlOut interface{}
			yamlErr := yaml.Unmarshal(output.Bytes(), &yamlOut)

			if yamlErr == nil {
				return nil, microerror.Mask(err)
			}

			matches := yamlErrorLinePattern.FindAllStringSubmatch(yamlErr.Error(), -1)
			if len(matches) == 0 {
				return nil, microerror.Mask(err)
			}

			lineNo, convErr := strconv.Atoi(matches[0][1])
			if convErr != nil {
				return nil, microerror.Mask(err)
			}
			lines := strings.Split(output.String(), "\n")

			fmt.Println(red(yamlErr.Error()))
			if lineNo > 1 {
				fmt.Println("> " + lines[lineNo-2])
			}
			fmt.Println("> " + red(lines[lineNo-1]))
			if lineNo < len(lines)-2 {
				fmt.Println("> " + lines[lineNo])
			}

			return nil, microerror.Mask(err)
		}

		pathList, err := svc.All()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, p := range pathList {
			paths[p] = true
		}
	}
	tf.values = values
	tf.paths = paths

	// fill in installation and app if possible
	{
		elements := strings.Split(filepath, "/")
		if strings.HasPrefix(filepath, "installations") {
			tf.installation = elements[1]
			tf.app = elements[3]
		} else if strings.HasPrefix(filepath, "default") {
			tf.app = elements[2]
		}
		// else it's an include file and has neither app nor installation
	}

	return tf, nil
}

func NormalPath(path string) string {
	if strings.HasPrefix(path, ".") {
		path = strings.TrimPrefix(path, ".")
	}
	return path
}

// includeExtract is a helper struct, with a method passed to template's
// funcmap. It collects filepaths used as arguments to "include" function in
// templates.
type includeExtract struct {
	Filepaths []string
}

func (ie *includeExtract) include(filepath string, data interface{}) string {
	filepath = "include/" + filepath + ".yaml.template"
	ie.Filepaths = append(ie.Filepaths, filepath)
	return ""
}

func (ie *includeExtract) clear() {
	ie.Filepaths = []string{}
}

func dummyFuncMap() template.FuncMap {
	// sprig.funcMap
	dummy := template.FuncMap{}
	for fName := range sprig.FuncMap() {
		dummy[fName] = func(args ...interface{}) string {
			return fName
		}
	}
	// built-ins, which might be affected by interface comparison
	for _, fName := range []string{"eq", "ne"} {
		dummy[fName] = func(args ...interface{}) string {
			return fName
		}
	}
	return dummy
}

package lint

import (
	"html/template"
	"log"
	"regexp"
	"strings"
	"text/template/parse"

	"github.com/Masterminds/sprig"

	"github.com/giantswarm/microerror"
	pathmodifier "github.com/giantswarm/valuemodifier/path"
)

var (
	fMap                = sprig.FuncMap()
	templatePathPattern = regexp.MustCompile(`(\.[a-zA-Z].[a-zA-Z0-9_\.]+)`)
)

func init() {
	fMap["include"] = func(f string, data interface{}) string {
		return ""
	}
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
	// overshadowing exactly the same path in other value files
	Overshadowing []*ValueFile
}

type TemplateFile struct {
	filepath     string
	installation string // optional for defaults
	app          string
	values       map[string]*TemplateValue
	staticValues map[string]interface{}

	sourceBytes    []byte
	sourceTemplate *template.Template
}

type TemplateValue struct {
	Path            string
	OccurrenceCount int
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
				Value:         value,
				UsedBy:        []*TemplateFile{},
				Overshadowing: []*ValueFile{},
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
	if !strings.HasSuffix(filepath, ".template") && !strings.HasSuffix(filepath, ".values.yaml.patch") {
		return nil, microerror.Maskf(executionFailedError, "given file is not a template: %q", filepath)
	}

	tf := &TemplateFile{
		filepath:    filepath,
		sourceBytes: body,
	}

	// extract values from template
	allValues := map[string]*TemplateValue{}
	{
		t, err := template.New(filepath).Funcs(fMap).Parse(string(body))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tf.sourceTemplate = t

		// todo: actually use plain yaml!
		plainYaml := ""
		for _, node := range t.Tree.Root.Nodes {
			if node.Type() == parse.NodeText {
				plainYaml += node.String()
				continue
			}

			nodePaths := templatePathPattern.FindAllString(node.String(), -1)
			for _, np := range nodePaths {
				normalPath := NormalPath(np)
				if _, ok := allValues[normalPath]; !ok {
					allValues[normalPath] = &TemplateValue{
						Path:            normalPath,
						OccurrenceCount: 1,
					}
				} else {
					allValues[normalPath].OccurrenceCount += 1
				}
			}
		}
		// TODO: KUBA
		log.Printf("KUBA: plainYaml: %q", plainYaml)
	}
	tf.values = allValues

	// fill in installation and app if possible
	{
		elements := strings.Split(filepath, "/")
		if strings.HasPrefix(filepath, "installations") {
			tf.installation = elements[1]
			tf.app = elements[3]
		} else if strings.HasPrefix(filepath, "default") {
			tf.app = elements[2]
		} else {
			return nil, microerror.Maskf(executionFailedError, "given file is not a template: %q", filepath)
		}
	}

	return tf, nil
}

func NormalPath(path string) string {
	if strings.HasPrefix(path, ".") {
		path = strings.TrimPrefix(path, ".")
	}
	return path
}

package template

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"text/template"

	"github.com/masterminds/sprig"
	"github.com/pkg/errors"
)

var (
	// ErrTemplateMissingContents is the error returned when a template
	// does not specify either a "content" argument, which is not
	// valid.
	ErrTemplateMissingContents = errors.New("template: must specify 'contents'")
)

// Template is the internal representation of an individual template to process.
// The template retains the relationship between it's contents and is
// responsible for it's own execution.
type Template struct {
	// contents is the string contents for the template.
	contents string

	// data used for templating from.
	data interface{}

	// leftDelim and rightDelim are the template delimiters.
	leftDelim  string
	rightDelim string

	// hexMD5 stores the hex version of the MD5
	hexMD5 string

	// errMissingKey causes the template processing to exit immediately if a map
	// is indexed with a key that does not exist.
	errMissingKey bool

	// functionDenylist are functions not permitted to be executed
	// when we render this template
	functionDenylist []string
}

// NewTemplateInput is used as input when creating the template.
type NewTemplateInput struct {
	// Contents are the raw template contents.
	Contents string

	// Data used for templating from.
	Data interface{}

	// ErrMissingKey causes the template parser to exit immediately with an error
	// when a map is indexed with a key that does not exist.
	ErrMissingKey bool

	// LeftDelim and RightDelim are the template delimiters.
	LeftDelim  string
	RightDelim string

	// FunctionDenylist are functions not permitted to be executed
	// when we render this template
	FunctionDenylist []string
}

// NewTemplate creates and parses a new Consul Template template at the given
// path. If the template does not exist, an error is returned. During
// initialization, the template is read and is parsed for dependencies. Any
// errors that occur are returned.
func NewTemplate(i *NewTemplateInput) (*Template, error) {
	if i == nil {
		i = &NewTemplateInput{}
	}

	if i.Contents == "" {
		return nil, ErrTemplateMissingContents
	}

	var t Template

	t.contents = i.Contents
	t.data = i.Data
	t.leftDelim = i.LeftDelim
	t.rightDelim = i.RightDelim
	t.errMissingKey = i.ErrMissingKey
	t.functionDenylist = i.FunctionDenylist

	// Compute the MD5, encode as hex
	hash := md5.Sum([]byte(t.contents))
	t.hexMD5 = hex.EncodeToString(hash[:])

	return &t, nil
}

// ID returns the identifier for this template.
func (t *Template) ID() string {
	return t.hexMD5
}

// ExecuteResult is the result of the template execution.
type ExecuteResult struct {
	// Output is the rendered result.
	Output []byte
}

// Execute evaluates this template in the provided context.
func (t *Template) Execute() (*ExecuteResult, error) {

	tmpl := template.New("")

	tmpl.Delims(t.leftDelim, t.rightDelim)

	tmpl.Funcs(funcMap(&funcMapInput{
		t:                tmpl,
		functionDenylist: t.functionDenylist,
	}))

	if t.errMissingKey {
		tmpl.Option("missingkey=error")
	} else {
		tmpl.Option("missingkey=zero")
	}

	tmpl, err := tmpl.Parse(t.contents)
	if err != nil {
		return nil, errors.Wrap(err, "parse")
	}

	// Execute the template into the writer
	var b bytes.Buffer
	if err := tmpl.Execute(&b, t.data); err != nil {
		return nil, errors.Wrap(err, "execute")
	}

	return &ExecuteResult{
		Output: b.Bytes(),
	}, nil
}

// funcMapInput is input to the funcMap, which builds the template functions.
type funcMapInput struct {
	t                *template.Template
	functionDenylist []string
}

// funcMap is the map of template functions to their respective functions.
func funcMap(i *funcMapInput) template.FuncMap {

	r := template.FuncMap{
		// Helper functions
		"parseYAML": parseYAML,
		"toYAML":    toYAML,
		// Debug functions
		"spew_dump":    spewDump,
		"spew_printf":  spewPrintf,
		"spew_sdump":   spewSdump,
		"spew_sprintf": spewSprintf,
	}

	sprigFuncs := sprig.FuncMap()
	// Removed these functions from the core Sprig package for security concerns
	delete(sprigFuncs, "env")
	delete(sprigFuncs, "expandenv")

	// add sprig functions
	for k, v := range sprigFuncs {
		r[k] = v
	}

	for _, bf := range i.functionDenylist {
		if _, ok := r[bf]; ok {
			r[bf] = denied
		}
	}

	return r
}

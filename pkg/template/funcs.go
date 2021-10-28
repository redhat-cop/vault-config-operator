package template

import (
	"bytes"

	spewLib "github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// parseYAML returns a structure for valid YAML
func parseYAML(s string) (interface{}, error) {
	if s == "" {
		return map[string]interface{}{}, nil
	}

	var data interface{}
	if err := yaml.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// toYAML converts the given structure into a deeply nested YAML string.
func toYAML(m map[string]interface{}) (string, error) {
	result, err := yaml.Marshal(m)
	if err != nil {
		return "", errors.Wrap(err, "toYAML")
	}
	return string(bytes.TrimSpace(result)), nil
}

// denied always returns an error, to be used in place of denied template functions
func denied(...string) (string, error) {
	return "", errors.New("function is disabled")
}

func spewSdump(args ...interface{}) (string, error) {
	return spewLib.Sdump(args...), nil
}

func spewSprintf(format string, args ...interface{}) (string, error) {
	return spewLib.Sprintf(format, args...), nil
}

func spewDump(args ...interface{}) (string, error) {
	spewLib.Dump(args...)
	return "", nil
}

func spewPrintf(format string, args ...interface{}) (string, error) {
	spewLib.Printf(format, args...)
	return "", nil
}

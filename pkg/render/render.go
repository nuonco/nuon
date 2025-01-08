package render

import (
	"bytes"
	"fmt"
	"text/template"

	"google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultPrefix string = "nuon"
)

func RenderString(inputVal string, intermediateData *structpb.Struct) (string, error) {
	if inputVal == "" {
		return "", nil
	}

	data := intermediateData.AsMap()

	temp, err := template.New("input").Option("missingkey=zero").Parse(inputVal)
	if err != nil {
		return inputVal, nil
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		return inputVal, fmt.Errorf("unable to execute template: %w", err)
	}

	outputVal := buf.String()
	if outputVal == "" {
		return "", fmt.Errorf("rendered value was empty, this usually means a bad interpolation config: %s", inputVal)
	}

	if outputVal == "<no value>" {
		return "", fmt.Errorf("rendered value was empty, which usually means a bad interpolation config: %s", inputVal)
	}
	return outputVal, nil
}

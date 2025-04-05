package render

import (
	"bytes"
	"fmt"
	"html/template"

	"google.golang.org/protobuf/types/known/structpb"
)

// This method is being deprecated, in favor of the more thorough render methods we have introduced
func Render(inputVal string, data map[string]interface{}) (string, error) {
	if inputVal == "" {
		return "", nil
	}

	_, isPrefixed := data[defaultPrefix]
	if !isPrefixed {
		data = map[string]interface{}{
			defaultPrefix: data,
		}
	}

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

func RenderString(inputVal string, intermediateData *structpb.Struct) (string, error) {
	return Render(inputVal, intermediateData.AsMap())
}

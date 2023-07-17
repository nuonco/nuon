package outputs

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"
	"google.golang.org/protobuf/types/known/structpb"
)

func TFOutputMetaToStructPB(tfom map[string]tfexec.OutputMeta) (*structpb.Struct, error) {
	msi := make(map[string]interface{})
	for key, outMeta := range tfom {
		msi[key] = outMeta.Value
	}
	if true {
		// round trip to json bytes to deal with RawMessage if necessary
		byts, err := json.Marshal(msi)
		if err != nil {
			return nil, fmt.Errorf("unable to convert to json: %w", err)
		}
		if err := json.Unmarshal(byts, &msi); err != nil {
			return nil, fmt.Errorf("unable to go from json back to a map[string]: %w", err)
		}
	}
	spb, err := structpb.NewStruct(msi)
	if err != nil {
		return nil, fmt.Errorf("TFOutputMetaToStructPB: error creating structpb: %w", err)
	}
	return spb, nil
}

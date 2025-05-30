package refs

import (
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func GetFakeRefs(refs []Ref) map[string]any {
	obj := make(map[string]any)

	for _, ref := range refs {
		ptr := obj
		refPieces := strings.Split(ref.Input, ".")
		if ref.Type == RefTypeActions || ref.Type == RefTypeComponents {
			refPieces = refPieces[4:]
		} else {
			refPieces = refPieces[3:]
		}

		for idx, piece := range refPieces {
			piece = strings.TrimSpace(piece)

			last := idx == len(refPieces)-1
			if last {
				ptr[piece] = generics.GetFakeObj[string]()
				continue
			}

			_, ok := ptr[piece]
			if !ok {
				ptr[piece] = make(map[string]any, 0)
			}
			ptr = ptr[piece].(map[string]any)
		}
	}

	return obj
}

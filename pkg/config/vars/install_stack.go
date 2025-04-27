package vars

import "github.com/powertoolsdev/mono/pkg/generics"

func (v *varsValidator) getInstallStack() installStackIntermediate {
	return generics.GetFakeObj[installStackIntermediate]()
}

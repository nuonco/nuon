package vars

import "github.com/nuonco/nuon/pkg/generics"

func (v *varsValidator) getInstallStack() installStackIntermediate {
	return generics.GetFakeObj[installStackIntermediate]()
}

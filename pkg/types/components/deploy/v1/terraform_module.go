package deployv1

import "strings"

func (t TerraformVersion) ToVersionString() string {
	if t == TerraformVersion_TERRAFORM_VERSION_LATEST {
		return "v1.4.5"
	}

	str := t.String()
	str = strings.ReplaceAll(str, "TERRAFORM_VERSION__", "v")
	str = strings.ReplaceAll(str, "_", ".")

	return str
}

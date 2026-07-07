package core

import "testing"

func TestDefaultMethodForCloud(t *testing.T) {
	for _, cloud := range []Cloud{CloudAWS, CloudGCP} {
		if got := DefaultMethodForCloud(cloud); got != MethodTerraform {
			t.Errorf("DefaultMethodForCloud(%q) = %q, want %q", cloud, got, MethodTerraform)
		}
	}
}

func TestValidateCloud(t *testing.T) {
	cases := []struct {
		cloud   Cloud
		wantErr bool
	}{
		{CloudAWS, false},
		{CloudGCP, false},
		{CloudAzure, true},
		{Cloud("bogus"), true},
	}
	for _, c := range cases {
		err := ValidateCloud(c.cloud)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateCloud(%q) err = %v, wantErr = %v", c.cloud, err, c.wantErr)
		}
	}
}

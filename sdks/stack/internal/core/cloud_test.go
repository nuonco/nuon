package core

import "testing"

func TestDefaultMethodForCloud(t *testing.T) {
	cases := map[Cloud]Method{
		CloudAWS: MethodSDK,
		CloudGCP: MethodTerraform,
	}
	for cloud, want := range cases {
		if got := DefaultMethodForCloud(cloud); got != want {
			t.Errorf("DefaultMethodForCloud(%q) = %q, want %q", cloud, got, want)
		}
	}
}

func TestValidateCloudMethod(t *testing.T) {
	cases := []struct {
		cloud   Cloud
		method  Method
		wantErr bool
	}{
		{CloudAWS, MethodSDK, false},
		{CloudAWS, MethodTerraform, false},
		{CloudGCP, MethodTerraform, false},
		{CloudGCP, MethodSDK, true},
		{CloudAzure, MethodTerraform, true},
		{Cloud("bogus"), MethodTerraform, true},
	}
	for _, c := range cases {
		err := ValidateCloudMethod(c.cloud, c.method)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateCloudMethod(%q, %q) err = %v, wantErr = %v", c.cloud, c.method, err, c.wantErr)
		}
	}
}

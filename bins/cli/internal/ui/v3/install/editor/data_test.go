package editor

import "testing"

func TestChangedInputs(t *testing.T) {
	mappings := []inputMapping{
		{name: "region", currentValue: "us-east-1"},
		{name: "license_key", sensitive: true},
		{name: "replicas", currentValue: "1"},
	}

	t.Run("blank sensitive field is omitted", func(t *testing.T) {
		got := changedInputs(mappings, []string{"us-west-2", "", "1"})
		if len(got) != 1 || got["region"] != "us-west-2" {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("typed sensitive field is sent", func(t *testing.T) {
		got := changedInputs(mappings, []string{"us-east-1", "new-key", "1"})
		if len(got) != 1 || got["license_key"] != "new-key" {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("cleared value is sent", func(t *testing.T) {
		got := changedInputs(mappings, []string{"", "", "1"})
		if len(got) != 1 || got["region"] != "" {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("nothing changed", func(t *testing.T) {
		got := changedInputs(mappings, []string{"us-east-1", "   ", "1"})
		if len(got) != 0 {
			t.Fatalf("got %#v", got)
		}
	})
}

package imageref

import "testing"

func TestImageRef(t *testing.T) {
	cases := []struct {
		name string
		src  Source
		want string
	}{
		{
			name: "digest only",
			src: Source{
				SourceImage:  "nginx",
				SourceDigest: "sha256:abcdef0123456789",
			},
			want: "nginx@sha256:abcdef0123456789",
		},
		{
			name: "registry-qualified repo",
			src: Source{
				SourceImage:  "ghcr.io/nuonco/foo",
				SourceDigest: "sha256:0011223344556677",
			},
			want: "ghcr.io/nuonco/foo@sha256:0011223344556677",
		},
		{
			name: "no digest returns empty",
			src: Source{
				SourceImage: "nginx",
				SourceRef:   "nginx:1.25.3",
				ResolvedTag: "1.25.3",
			},
			want: "",
		},
		{
			name: "no image returns empty",
			src: Source{
				SourceDigest: "sha256:abc",
			},
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ImageRef(tc.src); got != tc.want {
				t.Fatalf("ImageRef = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDisplayRef(t *testing.T) {
	cases := []struct {
		name string
		src  Source
		want string
	}{
		{
			name: "tag and digest, resolved tag preferred",
			src: Source{
				SourceImage:  "nginx",
				SourceRef:    "nginx:1.25.3",
				ResolvedTag:  "1.25.5",
				SourceDigest: "sha256:abcdef0123456789",
			},
			want: "nginx:1.25.5 (sha256:abcdef0)",
		},
		{
			name: "tag falls back to source ref tag",
			src: Source{
				SourceImage:  "nginx",
				SourceRef:    "nginx:1.25.3",
				SourceDigest: "sha256:abcdef0123456789",
			},
			want: "nginx:1.25.3 (sha256:abcdef0)",
		},
		{
			name: "no tag, digest only display",
			src: Source{
				SourceImage:  "nginx",
				SourceRef:    "nginx@sha256:abcdef0123456789",
				SourceDigest: "sha256:abcdef0123456789",
			},
			want: "nginx@sha256:abcdef0",
		},
		{
			name: "tag and image but no digest",
			src: Source{
				SourceImage: "nginx",
				ResolvedTag: "1.25.3",
			},
			want: "nginx:1.25.3",
		},
		{
			name: "legacy build only has source ref",
			src: Source{
				SourceRef: "nginx:1.25.3",
			},
			want: "nginx:1.25.3",
		},
		{
			name: "registry with port is not parsed as tag",
			src: Source{
				SourceImage:  "ghcr.io:5000/nuonco/foo",
				SourceRef:    "ghcr.io:5000/nuonco/foo:1.0",
				SourceDigest: "sha256:abcdef0123456789",
			},
			want: "ghcr.io:5000/nuonco/foo:1.0 (sha256:abcdef0)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DisplayRef(tc.src); got != tc.want {
				t.Fatalf("DisplayRef = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseTagFromRef(t *testing.T) {
	cases := []struct {
		ref  string
		want string
	}{
		{"nginx:1.25.3", "1.25.3"},
		{"nginx@sha256:abc", ""},
		{"nginx:1.0@sha256:abc", "1.0"},
		{"nginx", ""},
		{"ghcr.io/foo/bar:1.0", "1.0"},
		{"ghcr.io:5000/foo", ""},
		{"ghcr.io:5000/foo:1.0", "1.0"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := parseTagFromRef(tc.ref); got != tc.want {
			t.Errorf("parseTagFromRef(%q) = %q, want %q", tc.ref, got, tc.want)
		}
	}
}

func TestShortDigest(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"sha256:abcdef0123456789", "abcdef0"},
		{"sha256:abc", "abc"},
		{"abcdef0", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := shortDigest(tc.in); got != tc.want {
			t.Errorf("shortDigest(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSpecPullRef(t *testing.T) {
	const digest = "sha256:6b600461cdb1f1ffc311da10c80c423262f2c9edd9e5d74b519f20e004d49ef6"

	cases := []struct {
		name     string
		spec     Spec
		selected string
		want     string
	}{
		{"plain tag", Spec{Image: "nginx", Tag: "1.25.3"}, "", "1.25.3"},
		{"bare digest tag", Spec{Image: "nginx", Tag: digest}, "", digest},
		{"full ref in tag", Spec{Image: "nginx", Tag: "nginx@" + digest}, "", digest},
		{"digest baked in image, empty tag", Spec{Image: "nginx@" + digest}, "", digest},
		{"digest in image, tag still wins", Spec{Image: "nginx@" + digest, Tag: "1.25.3"}, "", "1.25.3"},
		{"empty tag, no digest", Spec{Image: "nginx"}, "", ""},
		{"registry path tag", Spec{Image: "ghcr.io/foo/bar", Tag: "1.0"}, "", "1.0"},
		{"update_policy uses selected tag", Spec{Image: "nginx", UpdatePolicy: "~1.25.0"}, "1.25.5", "1.25.5"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.spec.PullRef(tc.selected); got != tc.want {
				t.Errorf("Spec%+v.PullRef(%q) = %q, want %q", tc.spec, tc.selected, got, tc.want)
			}
		})
	}
}

func TestSpecIdentity(t *testing.T) {
	const digest = "sha256:6b600461cdb1f1ffc311da10c80c423262f2c9edd9e5d74b519f20e004d49ef6"

	cases := []struct {
		name         string
		spec         Spec
		selected     string
		wantImage    string
		wantRef      string
		wantResolved string
	}{
		{"plain tag", Spec{Image: "nginx", Tag: "1.25.3"}, "", "nginx", "nginx:1.25.3", "1.25.3"},
		{"bare digest tag", Spec{Image: "nginx", Tag: digest}, "", "nginx", "nginx@" + digest, ""},
		{"full ref in tag", Spec{Image: "nginx", Tag: "nginx@" + digest}, "", "nginx", "nginx@" + digest, ""},
		{"full ref in tag, empty image", Spec{Tag: "nginx@" + digest}, "", "nginx", "nginx@" + digest, ""},
		{"digest baked in image", Spec{Image: "nginx@" + digest}, "", "nginx", "nginx@" + digest, ""},
		{"update_policy records constraint", Spec{Image: "nginx", UpdatePolicy: "~1.25.0"}, "1.25.5", "nginx", "nginx:~1.25.0", "1.25.5"},
		{"registry path tag", Spec{Image: "ghcr.io/foo/bar", Tag: "1.0"}, "", "ghcr.io/foo/bar", "ghcr.io/foo/bar:1.0", "1.0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.spec.Identity(tc.selected)
			if got.SourceImage != tc.wantImage || got.SourceRef != tc.wantRef || got.ResolvedTag != tc.wantResolved {
				t.Errorf("Spec%+v.Identity(%q) = {Image:%q Ref:%q Resolved:%q}, want {Image:%q Ref:%q Resolved:%q}",
					tc.spec, tc.selected, got.SourceImage, got.SourceRef, got.ResolvedTag, tc.wantImage, tc.wantRef, tc.wantResolved)
			}
		})
	}
}

package sourceidentity_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/sourceidentity"
)

func sum256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func quote(value string) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func identityPreimage(i *sourceidentity.Identity) string {
	files := slices.Clone(i.Files)
	sort.Slice(files, func(a, b int) bool { return files[a].Path < files[b].Path })
	manifest := "null"
	if len(files) > 0 {
		entries := make([]string, len(files))
		for n, file := range files {
			executable := ""
			if file.Executable {
				executable = `,"executable":true`
			}
			entries[n] = fmt.Sprintf(`{"path":%s,"sha256":%s%s}`, quote(file.Path), quote(file.SHA256), executable)
		}
		manifest = "[" + strings.Join(entries, ",") + "]"
	}
	chartVersion := ""
	if i.ChartVersion != "" {
		chartVersion = `,"ChartVersion":` + quote(i.ChartVersion)
	}
	return fmt.Sprintf(`{"Version":%d,"ConfigChecksum":%s,"Kind":%s,"Root":%s,"Digest":%s,"Files":%s%s}`,
		i.Version, quote(i.ConfigChecksum), quote(i.Kind), quote(i.Root), quote(i.Digest), manifest, chartVersion)
}

func cloneIdentity(i *sourceidentity.Identity) *sourceidentity.Identity {
	if i == nil {
		return nil
	}
	out := *i
	out.Files = slices.Clone(i.Files)
	out.Reasons = slices.Clone(i.Reasons)
	return &out
}

func boundedText(maxRunes, maxBytes int) hegel.Generator[string] {
	// Text sizes count codepoints, not UTF-8 bytes.
	return hegel.Filter(hegel.Text().MinSize(0).MaxSize(maxRunes), func(value string) bool {
		return len(value) <= maxBytes
	})
}

func drawIdentity(t *hegel.T) *sourceidentity.Identity {
	i := &sourceidentity.Identity{
		Version: 1, ConfigChecksum: "config:" + hegel.Draw(t, boundedText(24, 48)),
		Kind:     hegel.Draw(t, hegel.SampledFrom([]string{"terraform", "helm", "kustomize", "pulumi", "image"})),
		Root:     hegel.Draw(t, hegel.SampledFrom([]string{"", "src", "nested/é"})),
		Coverage: "complete", Captured: true,
		Repository: "https://example.com/acme/repo", Commit: strings.Repeat("a", 40),
		Digest:       "sha256:" + sum256(hegel.Draw(t, boundedText(24, 48))),
		ChartVersion: hegel.Draw(t, hegel.SampledFrom([]string{"", "1.0.0", "1.2.3+meta"})),
	}
	count := hegel.Draw(t, hegel.Integers(0, 5))
	if hegel.Draw(t, hegel.Booleans()) {
		i.Files = make([]sourceidentity.File, 0, count)
	}
	for n := 0; n < count; n++ {
		i.Files = append(i.Files, sourceidentity.File{
			Path:       fmt.Sprintf("dir-%d/é-%d.txt", n%2, n),
			SHA256:     sum256(hegel.Draw(t, boundedText(32, 64))),
			Executable: hegel.Draw(t, hegel.Booleans()),
		})
	}
	i.Checksum = sum256(identityPreimage(i))
	return i
}

func TestIdentityKnownVectors(t *testing.T) {
	const emptyPreimage = `{"Version":1,"ConfigChecksum":"config","Kind":"terraform","Root":"src","Digest":"","Files":null}`
	const filePreimage = `{"Version":1,"ConfigChecksum":"config","Kind":"terraform","Root":"src","Digest":"","Files":[{"path":"main.tf","sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","executable":true}]}`
	for _, tc := range []struct {
		name, preimage, checksum string
		files                    []sourceidentity.File
	}{
		{"nil", emptyPreimage, "73c30970559780f0c3411e768ba3664c745fa3825fcdcfc4e5d1eeba44c29563", nil},
		{"empty", emptyPreimage, "73c30970559780f0c3411e768ba3664c745fa3825fcdcfc4e5d1eeba44c29563", []sourceidentity.File{}},
		{"file", filePreimage, "aea503a4ce04b5736bf16dee62bbca0ae7f3c77f54dab685e0ce84adb2a7f088", []sourceidentity.File{{Path: "main.tf", SHA256: sum256(""), Executable: true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := sourceidentity.New("terraform", "config")
			i.Root, i.Captured, i.Files = "src", true, tc.files
			require.Equal(t, tc.preimage, identityPreimage(i))
			require.Equal(t, tc.checksum, sum256(tc.preimage))
			i.Seal()
			require.Equal(t, tc.checksum, i.Checksum)
			require.NoError(t, i.Validate())
		})
	}
	for _, tc := range []struct {
		value any
		json  string
	}{
		{nil, "null"}, {[]string{}, "[]"}, {"<é>\n", `"\u003cé\u003e\n"`},
		{map[string]int{"z": 2, "a": 1}, `{"a":1,"z":2}`},
	} {
		require.Equal(t, sum256(tc.json), sourceidentity.Hash(tc.value))
	}
}

func TestIdentitySealProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		i := drawIdentity(t)
		want := i.Checksum
		for n := len(i.Files) - 1; n > 0; n-- {
			other := hegel.Draw(t, hegel.Integers(0, n))
			i.Files[n], i.Files[other] = i.Files[other], i.Files[n]
		}
		i.Seal()
		require.Equal(t, want, i.Checksum)
		require.NoError(t, i.Validate())
		require.Equal(t, want, i.EligibleChecksum())
		sealed := cloneIdentity(i)
		i.Seal()
		require.Equal(t, sealed, i)
		mutations := []func(*sourceidentity.Identity){
			func(j *sourceidentity.Identity) { j.Version++ },
			func(j *sourceidentity.Identity) { j.ConfigChecksum += "!" },
			func(j *sourceidentity.Identity) { j.Kind += "!" },
			func(j *sourceidentity.Identity) { j.Root += "/other" },
			func(j *sourceidentity.Identity) { j.Digest += "!" },
			func(j *sourceidentity.Identity) {
				j.Files = append(j.Files, sourceidentity.File{Path: "new.txt", SHA256: sum256("")})
			},
		}
		if len(i.Files) > 0 {
			mutations = append(mutations,
				func(j *sourceidentity.Identity) { j.Files[0].Path += "!" },
				func(j *sourceidentity.Identity) { j.Files[0].SHA256 = sum256(j.Files[0].SHA256) },
				func(j *sourceidentity.Identity) { j.Files[0].Executable = !j.Files[0].Executable },
				func(j *sourceidentity.Identity) { j.Files = j.Files[1:] },
			)
		}
		for n, mutate := range mutations {
			changed := cloneIdentity(i)
			mutate(changed)
			require.Error(t, changed.Validate(), "unsealed mutation %d", n)
			require.Empty(t, changed.EligibleChecksum())
			require.False(t, sourceidentity.Matches(i, changed))
			changed.Seal()
			require.Equal(t, sum256(identityPreimage(changed)), changed.Checksum)
			require.NotEqual(t, want, changed.Checksum, "committed field %d", n)
		}
	})
}

func TestIdentityEligibilityStateProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		base := drawIdentity(t)
		for _, version := range []int{0, 1, 2} {
			for _, coverage := range []struct {
				value   string
				reasons []string
				valid   bool
			}{
				{"complete", nil, true}, {"complete", []string{}, true},
				{"complete", []string{"uninspected"}, false},
				{"unknown", []string{"uninspected"}, true}, {"unknown", nil, false},
				{"unsupported", nil, false},
			} {
				for _, captured := range []bool{false, true} {
					for _, configPresent := range []bool{false, true} {
						for _, integrity := range []string{"sealed", "tampered", "missing"} {
							i := cloneIdentity(base)
							i.Version, i.Coverage, i.Reasons, i.Captured = version, coverage.value, coverage.reasons, captured
							if !configPresent {
								i.ConfigChecksum = ""
							}
							i.Checksum = sum256(identityPreimage(i))
							if integrity == "tampered" {
								i.Checksum = sum256(i.Checksum)
							} else if integrity == "missing" {
								i.Checksum = ""
							}
							before := cloneIdentity(i)
							valid := version == 1 && configPresent && coverage.valid && integrity == "sealed"
							wantEligible := ""
							if valid && captured && coverage.value == "complete" {
								wantEligible = i.Checksum
							}
							require.Equal(t, valid, i.Validate() == nil, "state: %+v", i)
							require.Equal(t, wantEligible, i.EligibleChecksum(), "state: %+v", i)
							require.Equal(t, valid, sourceidentity.Matches(i, cloneIdentity(i)), "receipt: %+v", i)
							require.Equal(t, before, i)
						}
					}
				}
			}
		}
	})
	var missing *sourceidentity.Identity
	require.Error(t, missing.Validate())
	require.Empty(t, missing.EligibleChecksum())
	require.False(t, sourceidentity.Matches(nil, nil))
}

func TestIdentityEvidenceProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		base := drawIdentity(t)
		if hegel.Draw(t, hegel.Booleans()) {
			base.Coverage, base.Reasons = "unknown", []string{"external_dependency"}
		}
		for name, mutate := range map[string]func(*sourceidentity.Identity){
			"repository": func(i *sourceidentity.Identity) { i.Repository += "/other" },
			"commit":     func(i *sourceidentity.Identity) { i.Commit = strings.Repeat("b", 40) },
			"capture":    func(i *sourceidentity.Identity) { i.Captured = false },
			"coverage": func(i *sourceidentity.Identity) {
				if i.Coverage == "complete" {
					i.Coverage, i.Reasons = "unknown", []string{"external_dependency"}
				} else {
					i.Coverage, i.Reasons = "complete", nil
				}
			},
		} {
			changed := cloneIdentity(base)
			mutate(changed)
			changed.Seal()
			require.Equal(t, base.Checksum, changed.Checksum, name)
			require.NoError(t, changed.Validate(), name)
			require.False(t, sourceidentity.Matches(base, changed), name)
			require.False(t, sourceidentity.Matches(changed, base), name)
			want := ""
			if changed.Coverage == "complete" && changed.Captured {
				want = changed.Checksum
			}
			require.Equal(t, want, changed.EligibleChecksum(), name)
		}
		unknown := cloneIdentity(base)
		unknown.Coverage, unknown.Reasons = "unknown", []string{"a"}
		other := cloneIdentity(unknown)
		other.Reasons[0] = "b"
		require.False(t, sourceidentity.Matches(unknown, other))
		require.True(t, sourceidentity.Matches(unknown, cloneIdentity(unknown)))
		require.Empty(t, unknown.EligibleChecksum())
		require.Empty(t, other.EligibleChecksum())
	})
}

func TestIdentityJSONProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		base := drawIdentity(t)
		for _, files := range [][]sourceidentity.File{nil, {}, base.Files} {
			for _, reasons := range [][]string{nil, {}, {"uninspected"}} {
				i := cloneIdentity(base)
				i.Files, i.Reasons = slices.Clone(files), slices.Clone(reasons)
				if len(reasons) > 0 {
					i.Coverage = "unknown"
				}
				i.Seal()
				before := cloneIdentity(i)
				data, err := json.Marshal(i)
				require.NoError(t, err)
				var out sourceidentity.Identity
				require.NoError(t, json.Unmarshal(data, &out))
				require.NoError(t, i.Validate())
				require.NoError(t, out.Validate())
				require.Equal(t, sum256(identityPreimage(i)), out.Checksum)
				wantEligible := ""
				if len(reasons) == 0 {
					wantEligible = out.Checksum
				}
				require.Equal(t, wantEligible, out.EligibleChecksum())
				require.True(t, sourceidentity.Matches(i, &out))
				require.True(t, sourceidentity.Matches(&out, i))
				require.Equal(t, before, i)
			}
		}
	})
}

func TestIdentityValidationProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		base := drawIdentity(t)
		base.Files = []sourceidentity.File{{Path: "file", SHA256: sum256("contents")}}
		for _, path := range []string{"", "/file", "..", "../file", "a/../../file", "a/../file", "./file", "a//file", "a/"} {
			i := cloneIdentity(base)
			i.Files[0].Path = path
			i.Checksum = sum256(identityPreimage(i))
			require.Error(t, i.Validate(), "path %q", path)
			require.Empty(t, i.EligibleChecksum())
			require.False(t, sourceidentity.Matches(i, cloneIdentity(i)))
		}
		for _, checksum := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.Repeat("g", 64)} {
			i := cloneIdentity(base)
			i.Files[0].SHA256 = checksum
			i.Checksum = sum256(identityPreimage(i))
			require.Error(t, i.Validate())
		}
		duplicate := cloneIdentity(base)
		duplicate.Files = append(duplicate.Files, duplicate.Files[0])
		duplicate.Checksum = sum256(identityPreimage(duplicate))
		require.Error(t, duplicate.Validate())
		for _, commit := range []string{"main", strings.Repeat("A", 40), strings.Repeat("a", 39), strings.Repeat("a", 41)} {
			i := cloneIdentity(base)
			i.Commit = commit
			i.Checksum = sum256(identityPreimage(i))
			require.Error(t, i.Validate())
			require.False(t, sourceidentity.IsCommit(commit))
		}
		image := cloneIdentity(base)
		image.Kind, image.Digest, image.Captured = "image", "latest", true
		image.Checksum = sum256(identityPreimage(image))
		require.Error(t, image.Validate())
		image.Captured = false
		require.NoError(t, image.Validate())
		before := cloneIdentity(base)
		base.Checksum = sum256(identityPreimage(base))
		before.Checksum = base.Checksum
		require.NoError(t, base.Validate())
		require.Equal(t, before, base)
	})
}

func TestIdentityManifestLimit(t *testing.T) {
	i := sourceidentity.New("terraform", "config")
	i.Files = make([]sourceidentity.File, 20000)
	for n := range i.Files {
		i.Files[n] = sourceidentity.File{Path: fmt.Sprintf("file-%05d", n), SHA256: sum256("")}
	}
	i.Checksum = sum256(identityPreimage(i))
	require.NoError(t, i.Validate())
	i.Files = append(i.Files, sourceidentity.File{Path: "extra", SHA256: sum256("")})
	i.Checksum = sum256(identityPreimage(i))
	require.Error(t, i.Validate())
}

func TestIdentityAliasingProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		i := drawIdentity(t)
		backing := make([]sourceidentity.File, 4)
		for n := range backing {
			backing[n] = sourceidentity.File{Path: fmt.Sprintf("file-%d", len(backing)-n), SHA256: sum256(fmt.Sprint(n))}
		}
		i.Files = backing[:3]
		i.Coverage = "unknown"
		reasons := []string{"a", "b", "unused", "sentinel"}
		i.Reasons = reasons[:2]
		i.Checksum = sum256(identityPreimage(i))
		other := *i
		if hegel.Draw(t, hegel.Booleans()) {
			other.Files = backing[1:]
		}
		other.Checksum = sum256(identityPreimage(&other))
		fileSnapshot, reasonSnapshot := slices.Clone(backing), slices.Clone(reasons)
		before, otherBefore := cloneIdentity(i), cloneIdentity(&other)
		for n := 0; n < 2; n++ {
			require.NoError(t, i.Validate())
			require.NoError(t, other.Validate())
			require.Empty(t, i.EligibleChecksum())
			require.Equal(t, before.Checksum == otherBefore.Checksum, sourceidentity.Matches(i, &other))
			require.True(t, sourceidentity.Matches(i, i))
			require.True(t, sourceidentity.Matches(i, cloneIdentity(i)))
		}
		require.Equal(t, before, i)
		require.Equal(t, otherBefore, &other)
		require.Equal(t, fileSnapshot, backing)
		require.Equal(t, reasonSnapshot, reasons)
	})
}

func TestIdentityUnknownReasonProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		i := drawIdentity(t)
		beforeChecksum := i.Checksum
		reasons := hegel.Draw(t, hegel.Lists(hegel.SampledFrom([]string{"a", "b", "c", "d"})).MinSize(1).MaxSize(10))
		want := slices.Clone(reasons)
		slices.Sort(want)
		want = slices.Compact(want)
		for _, reason := range reasons {
			i.MarkUnknown(reason)
		}
		require.Equal(t, "unknown", i.Coverage)
		require.Equal(t, want, i.Reasons)
		require.Equal(t, beforeChecksum, i.Checksum)
		require.NoError(t, i.Validate())
		require.Empty(t, i.EligibleChecksum())
	})
}

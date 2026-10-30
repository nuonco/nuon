package sourceidentity_test

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/pkg/sourceidentity"
)

func TestIdentityPairwiseReceiptAndReuseModel(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		contents := hegel.Draw(t, hegel.Maps(
			hegel.SampledFrom([]string{"a", "A", "nested/é", "nested/e\u0301", "space dir/file", "settings.json"}),
			hegel.Text().MaxSize(24)).MaxSize(6))
		var manifest []sourceidentity.File
		for _, path := range slices.Sorted(maps.Keys(contents)) {
			manifest = append(manifest, sourceidentity.File{Path: path, SHA256: sum256(contents[path]), Executable: hegel.Draw(t, hegel.Booleans())})
		}
		base := drawIdentity(t)
		base.Files = manifest
		states := []struct {
			version  int
			coverage string
			reasons  []string
			captured bool
			valid    bool
			eligible bool
		}{
			{1, "complete", nil, true, true, true},
			{1, "complete", []string{}, false, true, false},
			{1, "unknown", []string{"a", "b"}, true, true, false},
			{1, "unknown", []string{"b", "a"}, true, true, false},
			{1, "unknown", []string{"a", "b"}, false, true, false},
			{1, "unknown", nil, true, false, false},
			{1, "complete", []string{"a"}, true, false, false},
			{2, "complete", nil, true, false, false},
		}
		mutationCount := len(states) * len(states) * 3
		mutations := hegel.Draw(t, hegel.Lists(hegel.Booleans()).MinSize(mutationCount).MaxSize(mutationCount))
		for leftIndex, leftState := range states {
			for rightIndex, rightState := range states {
				flags := mutations[(leftIndex*len(states)+rightIndex)*3:][:3]
				left, right := cloneIdentity(base), cloneIdentity(base)
				left.Version, left.Coverage, left.Reasons, left.Captured = leftState.version, leftState.coverage, slices.Clone(leftState.reasons), leftState.captured
				right.Version, right.Coverage, right.Reasons, right.Captured = rightState.version, rightState.coverage, slices.Clone(rightState.reasons), rightState.captured
				if flags[0] {
					right.Repository += "/other"
				}
				if flags[1] {
					right.Commit = strings.Repeat("b", 40)
				}
				if flags[2] && len(right.Files) > 0 {
					right.Files[0].Executable = !right.Files[0].Executable
				}
				left.Checksum, right.Checksum = sum256(identityPreimage(left)), sum256(identityPreimage(right))
				leftBefore, rightBefore := cloneIdentity(left), cloneIdentity(right)
				wantReceipt := leftState.valid && rightState.valid && identityPreimage(left) == identityPreimage(right) &&
					left.Coverage == right.Coverage && slices.Equal(left.Reasons, right.Reasons) && left.Captured == right.Captured &&
					left.Repository == right.Repository && left.Commit == right.Commit
				require.Equal(t, wantReceipt, sourceidentity.Matches(left, right))
				require.Equal(t, wantReceipt, sourceidentity.Matches(right, left))
				wantReuse := leftState.eligible && rightState.eligible && identityPreimage(left) == identityPreimage(right)
				gotReuse := sourceidentity.Equivalent(left, right)
				require.Equal(t, wantReuse, gotReuse)
				require.Equal(t, leftBefore, left)
				require.Equal(t, rightBefore, right)
			}
		}
	})
}

func TestCaptureManifestLimitCountsCapturedFiles(t *testing.T) {
	t.Parallel()
	fs, repo := afero.NewMemMapFs(), "/repo"
	files := make([]diskFile, 19999)
	for n := range files {
		files[n] = diskFile{fmt.Sprintf("file-%05d.txt", n), "", 0644}
	}
	writeTree(t, fs, repo, "src", files)
	writeDiskFile(t, fs, repo, diskFile{"src/.git/ignored", "ignored", 0644})
	expected := sourceidentity.New("terraform", "config")
	expected.Root = "src"
	before := cloneIdentity(expected)
	source := rootedSource(t, fs, repo)
	for _, count := range []int{19999, 20000, 20001} {
		if count > len(files) {
			file := diskFile{fmt.Sprintf("file-%05d.txt", count-1), "", 0644}
			writeDiskFile(t, fs, repo, diskFile{"src/" + file.path, file.data, file.mode})
			files = append(files, file)
		}
		actual, err := sourceidentity.Capture(source, expected)
		if count <= 20000 {
			require.NoError(t, err)
			checkCapture(t, actual, expected, files, nil)
			require.Len(t, actual.Files, count)
		} else {
			require.ErrorContains(t, err, "file limit")
			require.Nil(t, actual)
		}
		require.Equal(t, before, expected)
	}
}

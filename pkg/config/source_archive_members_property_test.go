package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"

	temporaldataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
)

type archiveMemberModel struct {
	kind, name, path, text string
}

func drawArchiveMemberModel(t hegel.TestCase) []archiveMemberModel {
	var members []archiveMemberModel
	for _, fixed := range []archiveMemberModel{
		{kind: "metadata", name: "metadata", path: "metadata.toml"},
		{kind: "input", name: "inputs", path: "inputs.toml"},
		{kind: "policy", name: "policies", path: "policies.toml"},
		{kind: "break_glass", name: "break_glass", path: "break_glass.toml"},
		{kind: "operation_roles", name: "operation_roles", path: "operation_roles.toml"},
		{kind: "sandbox", name: "sandbox", path: "sandbox.toml"},
		{kind: "stack", name: "stack", path: "stack.toml"},
		{kind: "runner", name: "runner", path: "runner.toml"},
	} {
		if hegel.Draw(t, hegel.Booleans()) {
			fixed.text = hegel.Draw(t, sourceArchiveText())
			members = append(members, fixed)
		}
	}
	for _, role := range []string{"provision", "deprovision", "maintenance"} {
		if hegel.Draw(t, hegel.Booleans()) {
			members = append(members, archiveMemberModel{
				kind: "permission", name: role, path: "permissions/renamed_" + role + ".toml",
				text: fmt.Sprintf("type = %q\nname = %q\n", role, "{{ .nuon.install.id }}"),
			})
		}
	}
	count := hegel.Draw(t, hegel.Integers(0, 7))
	for i := 0; i < count; i++ {
		suffix := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,6}`, true))
		name, stem := fmt.Sprintf("logical_%d_%s", i, suffix), fmt.Sprintf("file_%d", i)
		entry := archiveMemberModel{name: name, text: fmt.Sprintf("name = %q\n", name)}
		switch hegel.Draw(t, hegel.Integers(0, 6)) {
		case 0:
			entry.kind, entry.path = "component", "components/"+stem+".toml"
		case 1:
			entry.kind, entry.path = "action", "actions/"+stem+".toml"
		case 2:
			entry.kind, entry.path = "runbook", "runbooks/"+stem+".toml"
		case 3:
			entry.kind, entry.path = "permission", "permissions/"+stem+".toml"
			entry.text += "type = \"custom\"\n"
		case 4:
			entry.kind, entry.path = "policy", "policies/"+stem+".toml"
		case 5:
			entry.kind, entry.path = "policy", "policies/"+stem+".toml"
			entry.name, entry.text = stem, "engine = \"opa\"\n"
		case 6:
			entry.kind, entry.path = "permission_policy", "permissions/policies/"+stem+".toml"
			entry.name = stem
		}
		if hegel.Draw(t, hegel.Booleans()) {
			entry.text = strings.ReplaceAll(entry.text, "\n", "\r\n")
		}
		members = append(members, entry)
	}
	for _, path := range []string{"permissions/boundary.json", "components/data.yaml", "scripts/empty.sh", "unknown/ignored.toml"} {
		members = append(members, archiveMemberModel{path: path, text: hegel.Draw(t, sourceArchiveText())})
	}
	return members
}

func assertSourceArchiveTransport(t require.TestingT, archive *SourceArchive) *SourceArchive {
	encoded, err := json.Marshal(archive)
	require.NoError(t, err)
	var direct SourceArchive
	require.NoError(t, json.Unmarshal(encoded, &direct))
	require.Equal(t, archive, &direct)
	converter := temporaldataconverter.NewJSONConverter()
	cfg := AppConfig{Version: "v2", SourceArchive: archive}
	payload, err := converter.ToPayload(cfg)
	require.NoError(t, err)
	var decoded AppConfig
	require.NoError(t, converter.FromPayload(payload, &decoded))
	require.Equal(t, archive, decoded.SourceArchive)
	intermediate, err := json.Marshal(cfg)
	require.NoError(t, err)
	withoutArchive, err := json.Marshal(AppConfig{Version: "v2"})
	require.NoError(t, err)
	require.Equal(t, string(withoutArchive), string(intermediate))
	return decoded.SourceArchive
}

func TestSourceArchiveMemberIndexAndTransportProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		model := drawArchiveMemberModel(t)
		for _, version := range []int{0, 1, 2, 3} {
			archive := NewSourceArchive()
			archive.SchemaVersion = version
			expectedFiles, expectedMembers := map[string]string{}, map[string]string{}
			for _, entry := range drawPermutation(t, model) {
				name := entry.name
				if version == 3 && entry.kind != "" {
					name = "stored_" + name
					entry.text = "invalid_toml = ["
				}
				expectedFiles[entry.path] = entry.text
				require.NoError(t, archive.AddFile(entry.path, []byte(entry.text)))
				if entry.kind == "" {
					continue
				}
				expectedMembers[entry.kind+":"+name] = entry.path
				if version == 3 {
					require.NoError(t, archive.AddMember(entry.kind, name, entry.path))
					require.NoError(t, archive.AddMember(entry.kind, name, entry.path))
					require.ErrorContains(t, archive.AddMember(entry.kind, name, "other/definition.toml"), "both define")
				}
			}
			if version < 3 {
				archive.Members["stale:identity"] = "missing.toml"
			}
			require.Equal(t, expectedFiles, archive.Files)
			require.NoError(t, archive.ReindexMembers())
			require.Equal(t, expectedMembers, archive.Members)
			require.NoError(t, archive.ReindexMembers())
			require.Equal(t, expectedMembers, archive.Members)
			require.Equal(t, expectedFiles, archive.Files)
			require.Equal(t, version, archive.SchemaVersion)
			require.NoError(t, archive.AddMember("", "ignored", "ignored.toml"))
			require.NoError(t, archive.AddMember("action", "", "ignored.toml"))
			require.Equal(t, expectedMembers, archive.Members)
			decoded := assertSourceArchiveTransport(t, archive)
			require.NoError(t, decoded.ReindexMembers())
			require.Equal(t, expectedMembers, decoded.Members)
			for _, key := range slices.Sorted(maps.Keys(expectedMembers)) {
				contents, ok := decoded.MemberSource(key)
				require.True(t, ok)
				require.Equal(t, expectedFiles[expectedMembers[key]], contents)
			}
			_, ok := decoded.MemberSource("missing:member")
			require.False(t, ok)
			reordered := &SourceArchive{SchemaVersion: version, Files: map[string]string{}, Members: map[string]string{}}
			for _, path := range drawPermutation(t, slices.Sorted(maps.Keys(expectedFiles))) {
				reordered.Files[path] = expectedFiles[path]
			}
			for _, key := range drawPermutation(t, slices.Sorted(maps.Keys(expectedMembers))) {
				reordered.Members[key] = expectedMembers[key]
			}
			require.NoError(t, reordered.ReindexMembers())
			require.Equal(t, archive, reordered)
			machine, err := json.Marshal(archive)
			require.NoError(t, err)
			reorderedMachine, err := json.Marshal(reordered)
			require.NoError(t, err)
			require.Equal(t, string(machine), string(reorderedMachine))
		}
	}, hegel.WithTestCases(40))
}

func TestSourceArchiveLegacyDuplicateIdentityProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		name := hegel.Draw(t, hegel.FromRegex(`[a-z]{1,8}`, true))
		for _, pair := range []struct {
			kind, directory, first, second string
		}{
			{"component", "components", fmt.Sprintf("name = %q\n", name), fmt.Sprintf("name = %q\n", name)},
			{"action", "actions", fmt.Sprintf("name = %q\n", name), fmt.Sprintf("name = %q\n", name)},
			{"runbook", "runbooks", fmt.Sprintf("name = %q\n", name), fmt.Sprintf("name = %q\n", name)},
			{"permission", "permissions", fmt.Sprintf("type = \"custom\"\nname = %q\n", name), fmt.Sprintf("name = %q\n", name)},
			{"permission", "permissions", "type = \"maintenance\"\nname = \"first\"\n", "type = \"maintenance\"\nname = \"second\"\n"},
			{"policy", "policies", "engine = \"opa\"\n", "name = \"a\"\n"},
		} {
			files := map[string]string{pair.directory + "/a.toml": pair.first, pair.directory + "/b.toml": pair.second}
			for _, version := range []int{0, 1, 2} {
				var previousError string
				for attempt := 0; attempt < 2; attempt++ {
					archive := &SourceArchive{SchemaVersion: version, Files: map[string]string{}, Members: map[string]string{"previous:index": "old.toml"}}
					for _, path := range drawPermutation(t, slices.Sorted(maps.Keys(files))) {
						archive.Files[path] = files[path]
					}
					err := archive.ReindexMembers()
					require.ErrorContains(t, err, "both define "+pair.kind+":")
					if attempt > 0 {
						require.Equal(t, previousError, err.Error())
					}
					previousError = err.Error()
					require.Equal(t, map[string]string{"previous:index": "old.toml"}, archive.Members)
					require.Equal(t, files, archive.Files)
				}
			}
			stored := &SourceArchive{SchemaVersion: 3, Files: files, Members: map[string]string{
				pair.kind + ":stored_a": pair.directory + "/a.toml",
				pair.kind + ":stored_b": pair.directory + "/b.toml",
			}}
			snapshot := maps.Clone(stored.Members)
			require.NoError(t, stored.ReindexMembers())
			require.Equal(t, snapshot, stored.Members)
		}
	}, hegel.WithTestCases(30))
}

func TestSourceArchiveIndexValidationProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		invalid := hegel.Draw(t, hegel.SampledFrom([]string{"name = [", "name = 42", "name = \"\"\n"}))
		for _, directory := range []string{"components", "actions", "runbooks", "permissions"} {
			path := directory + "/file.toml"
			legacy := &SourceArchive{SchemaVersion: hegel.Draw(t, hegel.Integers(0, 2)), Files: map[string]string{path: invalid}, Members: map[string]string{"previous:index": "old.toml"}}
			snapshot := maps.Clone(legacy.Members)
			require.Error(t, legacy.ReindexMembers())
			require.Equal(t, snapshot, legacy.Members)
			stored := &SourceArchive{SchemaVersion: 3, Files: maps.Clone(legacy.Files), Members: map[string]string{"action:stored": path}}
			require.NoError(t, stored.ReindexMembers())
			contents, ok := stored.MemberSource("action:stored")
			require.True(t, ok)
			require.Equal(t, invalid, contents)
		}
		for _, entry := range []struct {
			key, path, message string
		}{
			{"no_colon", "file.toml", "invalid member identity"},
			{":name", "file.toml", "invalid member identity"},
			{"action:", "file.toml", "invalid member identity"},
			{"action:name", "", "invalid member identity"},
			{"action:name", "missing.toml", "references missing file"},
		} {
			archive := &SourceArchive{SchemaVersion: 3, Files: map[string]string{"file.toml": ""}, Members: map[string]string{entry.key: entry.path}}
			snapshot := maps.Clone(archive.Members)
			require.ErrorContains(t, archive.ReindexMembers(), entry.message)
			require.Equal(t, snapshot, archive.Members)
		}
	}, hegel.WithTestCases(20))
}

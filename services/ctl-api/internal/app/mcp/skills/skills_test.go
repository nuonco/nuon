package skills

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestLoadedSkills(t *testing.T) {
	skills := All()
	require.NotEmpty(t, skills)

	for _, s := range skills {
		require.NotEmpty(t, s.Domain)
		require.NotEmpty(t, s.Name)
		require.NotEmpty(t, s.Description)
		require.NotEmpty(t, s.Body)
		require.Equal(t, "skill:///"+s.Domain+"/"+s.Name+".md", s.URI())
	}
}

func TestFind(t *testing.T) {
	s, ok := Find("apps", "write-readme")
	require.True(t, ok)
	require.Contains(t, s.Body, "Defensive rendering")

	_, ok = Find("apps", "does-not-exist")
	require.False(t, ok)
}

func TestIndex(t *testing.T) {
	idx := Index()
	require.Contains(t, idx, "apps/write-readme")
}

func TestParseSkillURI(t *testing.T) {
	domain, name, ok := parseSkillURI("skill:///apps/write-readme.md")
	require.True(t, ok)
	require.Equal(t, "apps", domain)
	require.Equal(t, "write-readme", name)

	_, _, ok = parseSkillURI("not-a-skill-uri")
	require.False(t, ok)
}

func TestReadIndexAndSkillResource(t *testing.T) {
	res, err := readIndex(context.Background(), &mcp.ReadResourceRequest{})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.True(t, strings.HasPrefix(res.Contents[0].Text, "# Skill index"))

	res, err = readSkill(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "skill:///apps/write-readme.md"},
	})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Contains(t, res.Contents[0].Text, "Defensive rendering")

	_, err = readSkill(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "skill:///apps/missing.md"},
	})
	require.Error(t, err)
}

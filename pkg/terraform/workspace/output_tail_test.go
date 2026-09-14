package workspace

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const diagnosticRecord = `{"@level":"error","@message":"Error: creating S3 Bucket (acme-artifacts): AccessDenied","diagnostic":{"summary":"creating S3 Bucket (acme-artifacts): AccessDenied","detail":"not authorized to perform: s3:CreateBucket"},"type":"diagnostic"}`

type failingTerraform struct {
	Terraform

	stream string
	err    error
}

func (f *failingTerraform) ApplyJSON(_ context.Context, w io.Writer, _ ...tfexec.ApplyOption) error {
	_, _ = io.WriteString(w, f.stream)
	return f.err
}

func (f *failingTerraform) DestroyJSON(_ context.Context, w io.Writer, _ ...tfexec.DestroyOption) error {
	_, _ = io.WriteString(w, f.stream)
	return f.err
}

func TestApply_ErrorCarriesOutputTail(t *testing.T) {
	client := &failingTerraform{
		stream: `{"@level":"info","@message":"Plan: 1 to add"}` + "\n" + diagnosticRecord + "\n",
		err:    errors.New("exit status 1"),
	}
	w := &workspace{v: validator.New(), root: t.TempDir()}

	_, err := w.apply(context.Background(), client, hclog.NewNullLogger())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error running apply: exit status 1")
	assert.Contains(t, err.Error(), "not authorized to perform: s3:CreateBucket")
}

func TestDestroy_ErrorCarriesOutputTail(t *testing.T) {
	client := &failingTerraform{
		stream: diagnosticRecord + "\n",
		err:    errors.New("exit status 1"),
	}
	w := &workspace{v: validator.New(), root: t.TempDir()}

	_, err := w.destroy(context.Background(), client, hclog.NewNullLogger())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error running destroy: exit status 1")
	assert.Contains(t, err.Error(), "AccessDenied")
}

func TestApplyDestroyPlan_ErrorCarriesOutputTail(t *testing.T) {
	client := &failingTerraform{
		stream: diagnosticRecord + "\n",
		err:    errors.New("exit status 1"),
	}
	w := &workspace{v: validator.New(), root: t.TempDir()}

	_, err := w.applyDestroyPlan(context.Background(), client, hclog.NewNullLogger())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error running destroy: exit status 1")
	assert.Contains(t, err.Error(), "AccessDenied")
}

func TestApply_ErrorWithNoOutputKeepsPlainMessage(t *testing.T) {
	client := &failingTerraform{err: errors.New("exit status 1")}
	w := &workspace{v: validator.New(), root: t.TempDir()}

	_, err := w.apply(context.Background(), client, hclog.NewNullLogger())

	require.Error(t, err)
	assert.Equal(t, "error running apply: exit status 1", err.Error())
}

func TestTailLines(t *testing.T) {
	for _, tc := range []struct {
		name     string
		in       string
		maxBytes int
		want     string
	}{
		{"empty", "", 16, ""},
		{"non-positive bound", "line\n", 0, ""},
		{"fits whole", "a\nb\nc\n", 32, "a\nb\nc"},
		{"cuts on line boundary", "aaaa\nbbbb\ncccc\n", 10, "bbbb\ncccc"},
		{"single oversized line", "aaaaaaaaaa", 4, "aaaa"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tailLines(tc.in, tc.maxBytes))
		})
	}
}

func TestTailLines_OversizedLineStaysValidUTF8(t *testing.T) {
	got := tailLines(strings.Repeat("€", 8), 5)

	assert.True(t, utf8.ValidString(got), "tail %q is not valid UTF-8", got)
	assert.Equal(t, "€", got)
}

func TestTailLines_BoundedByMaxBytes(t *testing.T) {
	in := strings.Repeat(diagnosticRecord+"\n", 200)

	got := tailLines(in, maxOutputTailBytes)

	assert.LessOrEqual(t, len(got), maxOutputTailBytes)
	assert.True(t, strings.HasPrefix(got, "{"), "tail should start at a record boundary, got %q", got[:1])
}

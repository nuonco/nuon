package teardown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	tests := map[string]struct {
		addr    string
		wantErr string
	}{
		"resource": {
			addr: "aws_rds_cluster.main",
		},
		"resource with integer index": {
			addr: "aws_instance.web[0]",
		},
		"resource with string index": {
			addr: `aws_s3_bucket.data["logs"]`,
		},
		"string index containing dots and brackets": {
			addr: `aws_s3_bucket.data["a.b[c]"]`,
		},
		"string index with escaped quote": {
			addr: `aws_s3_bucket.data["a\"b"]`,
		},
		"identifiers with dashes and digits": {
			addr: "aws_s3_bucket.data-bucket_2",
		},
		"module": {
			addr: "module.buckets",
		},
		"module instance": {
			addr: `module.buckets["primary"]`,
		},
		"nested modules": {
			addr: "module.storage[0].module.buckets",
		},
		"resource in module": {
			addr: `module.storage["a"].aws_s3_bucket.data[1]`,
		},
		"empty": {
			addr:    "",
			wantErr: "non-empty",
		},
		"surrounding whitespace": {
			addr:    " aws_rds_cluster.main",
			wantErr: "whitespace",
		},
		"data source": {
			addr:    "data.aws_ami.ubuntu",
			wantErr: "data source",
		},
		"data source in module": {
			addr:    "module.network.data.aws_vpc.main",
			wantErr: "data source",
		},
		"resource type only": {
			addr:    "aws_rds_cluster",
			wantErr: "expected a resource address",
		},
		"too many parts": {
			addr:    "aws_rds_cluster.main.id",
			wantErr: "expected a resource address",
		},
		"module keyword without name": {
			addr:    "module",
			wantErr: "module name",
		},
		"index on resource type": {
			addr:    "aws_instance[0].web",
			wantErr: "expected a resource address",
		},
		"unquoted string index": {
			addr:    "aws_s3_bucket.data[logs]",
			wantErr: "integer or a double-quoted string",
		},
		"unterminated string index": {
			addr:    `aws_s3_bucket.data["logs]`,
			wantErr: "unterminated",
		},
		"trailing dot": {
			addr:    "aws_rds_cluster.main.",
			wantErr: "invalid identifier",
		},
		"identifier starting with digit": {
			addr:    "aws_rds_cluster.1main",
			wantErr: "invalid identifier",
		},
		"wildcard": {
			addr:    "aws_s3_bucket.*",
			wantErr: "invalid identifier",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateAddress(tc.addr)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAddressMatches(t *testing.T) {
	tests := map[string]struct {
		entry  string
		target string
		want   bool
	}{
		"exact resource": {
			entry:  "aws_rds_cluster.main",
			target: "aws_rds_cluster.main",
			want:   true,
		},
		"resource without index selects every instance": {
			entry:  "aws_s3_bucket.data",
			target: `aws_s3_bucket.data["logs"]`,
			want:   true,
		},
		"resource with index selects only that instance": {
			entry:  `aws_s3_bucket.data["logs"]`,
			target: `aws_s3_bucket.data["assets"]`,
			want:   false,
		},
		"resource with index does not select unindexed resource": {
			entry:  "aws_instance.web[0]",
			target: "aws_instance.web",
			want:   false,
		},
		"integer and string indexes differ": {
			entry:  "aws_instance.web[0]",
			target: `aws_instance.web["0"]`,
			want:   false,
		},
		"name prefix is not a match": {
			entry:  "aws_s3_bucket.data",
			target: "aws_s3_bucket.data_logs",
			want:   false,
		},
		"different type with same name": {
			entry:  "aws_s3_bucket.data",
			target: "aws_s3_bucket_policy.data",
			want:   false,
		},
		"root resource does not select same resource in a module": {
			entry:  "aws_s3_bucket.data",
			target: "module.buckets.aws_s3_bucket.data",
			want:   false,
		},
		"module selects resources inside it": {
			entry:  "module.buckets",
			target: `module.buckets.aws_s3_bucket.data["logs"]`,
			want:   true,
		},
		"module selects nested module resources": {
			entry:  "module.storage",
			target: "module.storage[0].module.buckets.aws_s3_bucket.data",
			want:   true,
		},
		"module without index selects every instance": {
			entry:  "module.buckets",
			target: `module.buckets["primary"].aws_s3_bucket.data`,
			want:   true,
		},
		"module instance selects only that instance": {
			entry:  `module.buckets["primary"]`,
			target: `module.buckets["replica"].aws_s3_bucket.data`,
			want:   false,
		},
		"resource in unindexed module path selects every module instance": {
			entry:  "module.buckets.aws_s3_bucket.data",
			target: "module.buckets[1].aws_s3_bucket.data",
			want:   true,
		},
		"module name prefix is not a match": {
			entry:  "module.bucket",
			target: "module.buckets.aws_s3_bucket.data",
			want:   false,
		},
		"module does not select a root resource named like it": {
			entry:  "module.buckets",
			target: "aws_s3_bucket.buckets",
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			entry, err := parseAddress(tc.entry)
			require.NoError(t, err)
			target, err := parseAddress(tc.target)
			require.NoError(t, err)

			require.Equal(t, tc.want, entry.matches(target))
		})
	}
}

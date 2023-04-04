package eksclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_commands_Deprovision(t *testing.T) {
	cluster := generics.GetFakeObj[*ekstypes.Cluster]()
	clusterName := uuid.NewString()
	errGetCluster := fmt.Errorf("error getting cluster")

	tests := map[string]struct {
		eksClient   func(*gomock.Controller) awsEKSClient
		errExpected error
		assertFn    func(*testing.T, *ekstypes.Cluster)
	}{
		"happy path": {
			eksClient: func(mockCtl *gomock.Controller) awsEKSClient {
				client := NewMockawsEKSClient(mockCtl)
				req := &eks.DescribeClusterInput{
					Name: generics.ToPtr(clusterName),
				}
				resp := &eks.DescribeClusterOutput{
					Cluster: cluster,
				}

				client.EXPECT().DescribeCluster(gomock.Any(), req).
					Return(resp, nil)

				return client
			},
			assertFn: func(t *testing.T, resp *ekstypes.Cluster) {
				assert.Equal(t, cluster, resp)
			},
		},
		"error": {
			eksClient: func(mockCtl *gomock.Controller) awsEKSClient {
				client := NewMockawsEKSClient(mockCtl)
				req := &eks.DescribeClusterInput{
					Name: generics.ToPtr(clusterName),
				}
				client.EXPECT().DescribeCluster(gomock.Any(), req).
					Return(nil, errGetCluster)

				return client
			},
			errExpected: errGetCluster,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			awsEKSClient := test.eksClient(mockCtl)
			ec := &eksClient{
				ClusterName: clusterName,
			}

			cluster, err := ec.getCluster(ctx, awsEKSClient)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, cluster)
		})
	}
}

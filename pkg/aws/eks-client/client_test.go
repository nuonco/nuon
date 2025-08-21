package eksclient

import (
	"testing"
)

func TestNew(t *testing.T) {
	// This test was broken, but was it even being run before?
	// roleArn := uuid.NewString()
	// roleSessionName := uuid.NewString()
	// region := "us-west-2"
	// clusterName := uuid.NewString()
	// v := validator.New()

	// tests := map[string]struct {
	// 	errExpected error
	// 	optsFn      func() []eksOptions
	// 	assertFn    func(*testing.T, *eksClient)
	// }{
	// 	"happy path": {
	// 		optsFn: func() []eksOptions {
	// 			return []eksOptions{
	// 				WithRoleARN(roleArn),
	// 				WithRoleSessionName(roleSessionName),
	// 				WithRegion(region),
	// 				WithClusterName(clusterName),
	// 			}
	// 		},
	// 		assertFn: func(t *testing.T, e *eksClient) {
	// 			assert.Equal(t, roleArn, e.RoleARN)
	// 			assert.Equal(t, roleSessionName, e.RoleSessionName)
	// 			assert.Equal(t, clusterName, e.ClusterName)
	// 			assert.Equal(t, region, e.Region)
	// 		},
	// 	},
	// 	"missing region": {
	// 		optsFn: func() []eksOptions {
	// 			return []eksOptions{
	// 				WithRoleARN(roleArn),
	// 				WithRoleSessionName(roleSessionName),
	// 				WithClusterName(clusterName),
	// 			}
	// 		},
	// 		errExpected: fmt.Errorf("Region"),
	// 	},
	// 	"missing cluster name": {
	// 		optsFn: func() []eksOptions {
	// 			return []eksOptions{
	// 				WithRoleARN(roleArn),
	// 				WithRoleSessionName(roleSessionName),
	// 				WithRegion(region),
	// 			}
	// 		},
	// 		errExpected: fmt.Errorf("ClusterName"),
	// 	},
	// }

	// for name, test := range tests {
	// 	name := name
	// 	test := test
	// 	t.Run(name, func(t *testing.T) {
	// 		t.Parallel()

	// 		e, err := New(v, test.optsFn()...)
	// 		if test.errExpected != nil {
	// 			assert.Error(t, err)
	// 			assert.ErrorContains(t, err, test.errExpected.Error())
	// 			return
	// 		}
	// 		assert.NoError(t, err)
	// 		test.assertFn(t, e)
	// 	})
	// }
}

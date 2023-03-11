package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	rbacv1 "k8s.io/api/rbac/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacapplyv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
)

type testK8sRoleBindingCreator struct {
	mock.Mock
}

func (t *testK8sRoleBindingCreator) Apply(ctx context.Context, req *rbacapplyv1.RoleBindingApplyConfiguration, opts apimetav1.ApplyOptions) (*rbacv1.RoleBinding, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*rbacv1.RoleBinding), args.Error(1)
	}

	return nil, args.Error(1)
}

func getFakeCreateRoleBindingRequest() CreateRoleBindingRequest {
	id := uuid.NewString()
	return CreateRoleBindingRequest{
		InstallID:            id,
		TokenSecretNamespace: "default",
		OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
		NamespaceName:        id,
	}
}

func Test_roleBindingCreatorImpl_createRoleBinding(t *testing.T) {
	errRoleBindingCreate := fmt.Errorf("error creating role binding")
	req := getFakeCreateRoleBindingRequest()

	tests := map[string]struct {
		clientFn    func() k8sRoleBindingCreator
		assertFn    func(*testing.T, k8sRoleBindingCreator)
		errExpected error
	}{
		"happy path": {
			clientFn: func() k8sRoleBindingCreator {
				client := &testK8sRoleBindingCreator{}
				client.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(&rbacv1.RoleBinding{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client k8sRoleBindingCreator) {
				obj := client.(*testK8sRoleBindingCreator)
				obj.AssertNumberOfCalls(t, "Apply", 1)

				cr := obj.Calls[0].Arguments[1].(*rbacapplyv1.RoleBindingApplyConfiguration)
				assert.NotNil(t, cr)

				// ensure subject is configured correctly
				assert.Equal(t, "ServiceAccount", *cr.Subjects[0].Kind)
				assert.Equal(t, req.InstallID, *cr.Subjects[0].Namespace)
				assert.Equal(t, runnerServiceAccountName(req.InstallID), *cr.Subjects[0].Name)

				// ensure role is configured correctly
				assert.Equal(t, "ClusterRole", *cr.RoleRef.Kind)
				assert.Equal(t, "edit", *cr.RoleRef.Name)
			},
			errExpected: nil,
		},
		"error returned": {
			clientFn: func() k8sRoleBindingCreator {
				client := &testK8sRoleBindingCreator{}
				client.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(&rbacv1.RoleBinding{}, errRoleBindingCreate)
				return client
			},
			assertFn: func(t *testing.T, client k8sRoleBindingCreator) {
				obj := client.(*testK8sRoleBindingCreator)
				obj.AssertNumberOfCalls(t, "Apply", 1)

				req := obj.Calls[0].Arguments[1].(*rbacapplyv1.RoleBindingApplyConfiguration)
				assert.NotNil(t, req)
			},
			errExpected: errRoleBindingCreate,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.clientFn()
			r := roleBindingCreatorImpl{}

			err := r.createRoleBinding(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			} else {
				assert.Nil(t, err)
			}

			test.assertFn(t, client)
		})
	}
}

package sandbox

import "testing"

// TODO(jm): readd unit tests once this is stabilized
func Test_planCreatorImpl_createPlan(t *testing.T) {
	//req := generics.GetFakeObj[*planactivitiesv1.CreatePlanRequest]()
	//longIDs := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	//shortIDs, err := shortid.ParseStrings(longIDs...)
	//assert.NoError(t, err)

	//req.OrgID = shortIDs[0]
	//req.AppID = shortIDs[1]
	//req.DeploymentID = shortIDs[2]
	//req.Config.WaypointTokenSecretTemplate = "token-%s"
	//assert.NoError(t, req.validate())

	//errCreatePlan := fmt.Errorf("err creating plan")
	//tests := map[string]struct {
	//builderFn   func() *mockBuilder
	//assertFn    func(*testing.T, *planv1.WaypointPlan)
	//errExpected error
	//}{
	//"happy path - metadata": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//meta := plan.Metadata
	//assert.NotNil(t, meta)
	//assert.Equal(t, longIDs[0], meta.OrgId)
	//assert.Equal(t, shortIDs[0], meta.OrgShortId)

	//assert.Equal(t, longIDs[1], meta.AppId)
	//assert.Equal(t, shortIDs[1], meta.AppShortId)

	//assert.Equal(t, longIDs[2], meta.DeploymentId)
	//assert.Equal(t, shortIDs[2], meta.DeploymentShortId)
	//},
	//},
	//"happy path - waypoint server": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//wpPlan := plan.WaypointServer
	//cfg := req.Config

	//expectedAddr := client.DefaultOrgServerAddress(cfg.WaypointServerRootDomain, req.OrgID)
	//assert.Equal(t, expectedAddr, wpPlan.Address)

	//expectedTokenSecretName := fmt.Sprintf(cfg.WaypointTokenSecretTemplate, req.OrgID)
	//assert.Equal(t, expectedTokenSecretName, wpPlan.TokenSecretName)
	//assert.Equal(t, cfg.WaypointTokenSecretNamespace, wpPlan.TokenSecretNamespace)
	//},
	//},
	//"happy path - ecr repository": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//ecrPlan := plan.EcrRepositoryRef
	//cfg := req.Config

	//expectedRepoName := fmt.Sprintf("%s/%s", req.OrgID, req.AppID)
	//expectedRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", cfg.OrgsECRRegistryID,
	//cfg.OrgsECRRegion, expectedRepoName)
	//expectedRepoArn := fmt.Sprintf("%s/%s", cfg.OrgsECRRegistryARN, expectedRepoName)

	//assert.Equal(t, cfg.OrgsECRRegistryID, ecrPlan.RegistryId)
	//assert.Equal(t, expectedRepoName, ecrPlan.RepositoryName)
	//assert.Equal(t, expectedRepoArn, ecrPlan.RepositoryArn)
	//assert.Equal(t, expectedRepoURI, ecrPlan.RepositoryUri)
	//assert.Equal(t, req.DeploymentID, ecrPlan.Tag)
	//assert.Equal(t, cfg.OrgsECRRegion, ecrPlan.Region)
	//},
	//},
	//"happy path - waypoint ref": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//wpPlan := plan.WaypointRef

	//assert.Equal(t, req.OrgID, wpPlan.Project)
	//assert.Equal(t, req.AppID, wpPlan.Workspace)
	//assert.Equal(t, req.Component.Name, wpPlan.App)
	//assert.Contains(t, wpPlan.SingletonId, req.Component.Name)
	//assert.Contains(t, wpPlan.SingletonId, req.DeploymentID)
	//assert.Equal(t, req.OrgID, wpPlan.RunnerId)
	//assert.Equal(t, req.OrgID, wpPlan.OnDemandRunnerConfig)
	//assert.Equal(t, defaultJobTimeoutSeconds, wpPlan.JobTimeoutSeconds)

	//assert.Equal(t, req.OrgID, wpPlan.Labels["org-id"])
	//assert.Equal(t, req.DeploymentID, wpPlan.Labels["deployment-id"])
	//assert.Equal(t, req.AppID, wpPlan.Labels["app-id"])
	//assert.Equal(t, req.Component.Name, wpPlan.Labels["component-name"])

	//assert.Equal(t, "waypoint-hcl", wpPlan.HclConfig)
	//assert.Equal(t, waypointv1.Hcl_HCL.String(), wpPlan.HclConfigFormat)
	//},
	//},
	//"happy path - outputs": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//oPlan := plan.Outputs
	//cfg := req.Config

	//assert.Equal(t, cfg.DeploymentsBucket, oPlan.Bucket)
	//assert.Equal(t, req.DeploymentsBucketPrefix, oPlan.BucketPrefix)
	//assert.Equal(t, req.DeploymentsBucketAssumeRoleARN, oPlan.BucketAssumeRoleArn)

	//assert.Contains(t, oPlan.LogsKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.LogsKey, "logs.txt")

	//assert.Contains(t, oPlan.EventsKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.EventsKey, "events.json")

	//assert.Contains(t, oPlan.ArtifactKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.ArtifactKey, "artifacts.json")
	//},
	//},
	//"happy path - component": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//assert.True(t, proto.Equal(req.Component, plan.Component))
	//},
	//},
	//"error - builder": {
	//builderFn: func() *mockBuilder {
	//obj := &mockBuilder{}
	//obj.On("Render").Return(nil, waypointv1.Hcl_HCL, errCreatePlan)
	//obj.On("WithComponent", mock.Anything).Return()
	//obj.On("WithMetadata", mock.Anything).Return()
	//obj.On("WithECRRef", mock.Anything).Return()
	//return obj
	//},
	//errExpected: errCreatePlan,
	//},
	//}

	//for name, test := range tests {
	//t.Run(name, func(t *testing.T) {
	//pc := planCreatorImpl{}
	//builder := test.builderFn()

	//buildPlan, err := pc.createPlan(req, builder)
	//if test.errExpected != nil {
	//assert.ErrorContains(t, err, test.errExpected.Error())
	//return
	//}
	//assert.NoError(t, err)
	//test.assertFn(t, buildPlan)
	//})
	//}
}

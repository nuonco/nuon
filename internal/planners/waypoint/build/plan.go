package build

import (
	"context"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

func (p *planner) GetPlan(ctx context.Context) (*planv1.WaypointPlan, error) {
	//ecrRepoName := fmt.Sprintf("%s/%s", req.OrgId, req.AppId)
	//ecrRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", req.Config.OrgsECRRegistryID,
	//req.Config.OrgsECRRegion, ecrRepoName)

	//plan := &planv1.WaypointPlan{
	//Metadata: &planv1.Metadata{
	//OrgId:	     req.OrgId,
	//OrgShortId:	     req.OrgShortId,
	//AppId:	     req.AppId,
	//AppShortId:	     req.AppShortId,
	//DeploymentId:      req.DeploymentId,
	//DeploymentShortId: req.DeploymentShortId,
	//},
	//WaypointServer: &planv1.WaypointServerRef{
	//Address:		client.DefaultOrgServerAddress(req.Config.WaypointServerRootDomain, req.OrgID),
	//TokenSecretName:	fmt.Sprintf(req.Config.WaypointTokenSecretTemplate, req.OrgID),
	//TokenSecretNamespace: req.Config.WaypointTokenSecretNamespace,
	//},
	//EcrRepositoryRef: &planv1.ECRRepositoryRef{
	//RegistryId:	  req.Config.OrgsECRRegistryID,
	//RepositoryName: ecrRepoName,
	//RepositoryArn:  fmt.Sprintf("%s/%s", req.Config.OrgsECRRegistryARN, ecrRepoName),
	//RepositoryUri:  ecrRepoURI,
	//Tag:		  req.DeploymentID,
	//Region:	  req.Config.OrgsECRRegion,
	//},
	//WaypointRef: &planv1.WaypointRef{
	//Project:     req.OrgID,
	//Workspace:   req.AppID,
	//App:	       req.Component.Name,
	//SingletonId: fmt.Sprintf("%s-%s", req.DeploymentID, req.Component.Name),
	//Labels: map[string]string{
	//"deployment-id":  req.DeploymentID,
	//"app-id":	    req.AppID,
	//"org-id":	    req.OrgID,
	//"component-name": req.Component.Name,
	//},
	//RunnerId:		req.OrgID,
	//OnDemandRunnerConfig: req.OrgID,
	//JobTimeoutSeconds:	defaultJobTimeoutSeconds,
	//},
	//Outputs: &planv1.Outputs{
	//Bucket:	       req.Config.DeploymentsBucket,
	//BucketPrefix:        req.DeploymentsBucketPrefix,
	//BucketAssumeRoleArn: req.DeploymentsBucketAssumeRoleARN,

	//// TODO(jm): these aren't being used until we've fully implemented the executor
	//LogsKey:     filepath.Join(req.DeploymentsBucketPrefix, "logs.txt"),
	//EventsKey:   filepath.Join(req.DeploymentsBucketPrefix, "events.json"),
	//ArtifactKey: filepath.Join(req.DeploymentsBucketPrefix, "artifacts.json"),
	//},
	//Component: req.Component,
	//}

	//// configure builder
	//builder.WithComponent(req.Component)
	//builder.WithMetadata(plan.Metadata)
	//builder.WithECRRef(plan.EcrRepositoryRef)
	//cfg, cfgFmt, err := builder.Render()
	//if err != nil {
	//return nil, fmt.Errorf("unable to render config: %w", err)
	//}
	//plan.WaypointRef.HclConfig = string(cfg)
	//plan.WaypointRef.HclConfigFormat = cfgFmt.String()
	//return plan, nil
	return nil, nil
}

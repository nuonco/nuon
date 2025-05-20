package orgiam

func ProvisionIAMCallback(req *ProvisionIAMRequest) string {
	return req.WorkflowID
}

func DeprovisionIAMCallback(req *DeprovisionIAMRequest) string {
	return req.WorkflowID
}

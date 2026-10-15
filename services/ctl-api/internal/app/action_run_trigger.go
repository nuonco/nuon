package app

func BuildActionRunTrigger(run *InstallActionWorkflowRun, wf *Workflow, rbRun *InstallRunbookRun) *ActionRunTrigger {
	if run.TriggerType == ActionWorkflowTriggerTypeCron {
		return &ActionRunTrigger{Type: ActionRunTriggerTypeCron}
	}

	if rbRun != nil || run.TriggeredByType == "runbook" {
		trigger := &ActionRunTrigger{Type: ActionRunTriggerTypeRunbook}
		if rbRun != nil {
			trigger.Actor = AccountOrNil(rbRun.CreatedBy)
			trigger.ParentType = "runbook_runs"
			trigger.ParentID = rbRun.ID
			trigger.ParentName = rbRun.InstallRunbook.Runbook.Name
		} else if wf != nil {
			trigger.Actor = AccountOrNil(wf.CreatedBy)
		}
		return trigger
	}

	trigger := &ActionRunTrigger{Type: ActionRunTriggerTypeFor(run)}
	switch {
	case wf != nil:
		trigger.Actor = AccountOrNil(wf.CreatedBy)
		trigger.ParentType = string(wf.Type)
		trigger.ParentID = wf.ID
		trigger.ParentName = wf.Name
	case run.TriggeredByType == "account":
		trigger.Actor = AccountOrNil(run.CreatedBy)
	default:
		trigger.Type = ActionRunTriggerTypeSystem
	}

	return trigger
}

func ActionRunTriggerTypeFor(run *InstallActionWorkflowRun) ActionRunTriggerType {
	switch run.TriggerType {
	case ActionWorkflowTriggerTypeManual:
		return ActionRunTriggerTypeManual
	case ActionWorkflowTriggerTypeAdHoc:
		return ActionRunTriggerTypeAdhoc
	case ActionWorkflowTriggerTypePreDeployComponent,
		ActionWorkflowTriggerTypePostDeployComponent,
		ActionWorkflowTriggerTypePreTeardownComponent,
		ActionWorkflowTriggerTypePostTeardownComponent:
		return ActionRunTriggerTypeDeploy
	default:
		return ActionRunTriggerTypeWorkflow
	}
}

func AccountOrNil(acct Account) *Account {
	if acct.ID == "" {
		return nil
	}
	return &acct
}

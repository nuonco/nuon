import { WorkflowTimeline } from '../components/organisms/WorkflowTimeline'
import { useInstall } from '../providers/install-provider'
import { useOrg } from '../providers/org-provider'
import { useInstallPageChrome } from './InstallLayout'

export const InstallActivity = () => {
  useInstallPageChrome('Activity')

  const { orgId } = useOrg()
  const { install, installId } = useInstall()

  const driftedWorkflowIds = new Set(
    (install?.drifted_objects ?? [])
      .map((object) => object?.install_workflow_id)
      .filter((id): id is string => !!id)
  )

  return (
    <WorkflowTimeline
      orgId={orgId}
      owner={{ kind: 'install', installId: installId ?? '' }}
      driftedWorkflowIds={driftedWorkflowIds}
    />
  )
}

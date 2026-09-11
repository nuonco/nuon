import { WorkflowTimeline } from '../components/organisms/WorkflowTimeline'
import { useApp } from '../providers/app-provider'
import { useAppBranch } from '../providers/app-branch-provider'
import { useOrg } from '../providers/org-provider'
import { useAppBranchPageChrome } from './AppBranchLayout'

export const AppBranchActivity = () => {
  useAppBranchPageChrome('Activity')

  const { orgId } = useOrg()
  const { appId } = useApp()
  const { branchId } = useAppBranch()

  return (
    <WorkflowTimeline
      orgId={orgId}
      owner={{ kind: 'app', appId: appId ?? '', branchId: branchId ?? '' }}
    />
  )
}

import { useBranchConfigSections } from '@/components/branches/BranchRunChangesSummary'
import {
  TemplateChangesContent,
  configChangesBaseline,
} from '@/components/branches/ConfigChanges'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

interface IPullRequestConfigChanges {
  branchId: string
  appBranchRunId: string
}

export const PullRequestConfigChanges = ({
  branchId,
  appBranchRunId,
}: IPullRequestConfigChanges) => {
  const { org } = useOrg()
  const { app } = useApp()
  const config = useBranchConfigSections({ branchId, appBranchRunId })

  return (
    <TemplateChangesContent
      sections={config.sections}
      previousSha={config.previousSha}
      sha={config.sha}
      baseline={configChangesBaseline(config.comparison, {
        orgId: org?.id,
        appId: app?.id,
        branchId,
      })}
      isLoading={config.isLoading}
      isError={config.isError}
    />
  )
}

import { useBranchConfigSections } from '@/components/branches/BranchRunChangesSummary'
import { TemplateChangesContent } from '@/components/branches/ConfigChanges'

interface IPullRequestConfigChanges {
  branchId: string
  appBranchRunId: string
}

export const PullRequestConfigChanges = ({
  branchId,
  appBranchRunId,
}: IPullRequestConfigChanges) => {
  const config = useBranchConfigSections({ branchId, appBranchRunId })

  return (
    <TemplateChangesContent
      sections={config.sections}
      previousSha={config.previousSha}
      sha={config.sha}
      isLoading={config.isLoading}
      isError={config.isError}
    />
  )
}

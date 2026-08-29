import { useMemo } from 'react'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { PreviewConfigSection } from '@/components/branches/PreviewConfigSection'
import { latestBranchConfig } from '@/utils/branch-utils'
import { BranchTabPage } from './BranchTabPage'

export const BranchPreviewTab = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch, refresh } = useBranch()
  const orgId = org.id!
  const appId = app.id!
  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])

  return (
    <BranchTabPage
      tab="Preview"
      heading="Preview"
      subheading="Default settings for preview runs on this branch."
    >
      <PreviewConfigSection
        branch={branch}
        currentConfig={currentConfig}
        orgId={orgId}
        appId={appId}
        onSuccess={refresh}
      />
    </BranchTabPage>
  )
}

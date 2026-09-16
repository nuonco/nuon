import { useMemo } from 'react'
import { useParams } from 'react-router'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { BranchProvider } from '@/providers/branch-provider'
import { BranchCISettings } from '@/components/branches/BranchCISettings'
import { BranchSourceCard } from '@/components/branches/BranchSourceCard'
import {
  EditBranchNameButton,
  EditBranchNameModal,
} from '@/components/branches/EditBranchNameModal'
import { DeleteBranchModal } from '@/components/branches/BranchDetailActions'
import { PreviewConfigSection } from '@/components/branches/PreviewConfigSection'
import { latestBranchConfig } from '@/utils/branch-utils'

const BranchSettingsCards = () => {
  const { app } = useApp()
  const { org } = useOrg()
  const { branch, refresh } = useBranch()
  const { addModal } = useSurfaces()
  const appId = app.id!

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])
  const hasGitHubSource =
    !!currentConfig?.connected_github_vcs_config ||
    !!currentConfig?.public_git_vcs_config

  return (
    <div className="flex flex-col gap-4">
      <Card className="gap-4 p-4">
        <div className="flex items-center justify-between gap-2">
          <div className="flex flex-col gap-0.5">
            <Text weight="strong">Branch name</Text>
            <span className="flex items-center gap-2">
              <Text variant="subtext" theme="neutral">
                {branch.name}
              </Text>
              {branch.managed_by ? (
                <LabelBadge
                  labelKey="managed by"
                  labelValue={branch.managed_by}
                  size="sm"
                  theme={branch.managed_by === 'config' ? 'brand' : 'default'}
                />
              ) : null}
            </span>
          </div>
          <EditBranchNameButton branch={branch} onSuccess={refresh} />
        </div>
      </Card>

      <BranchSourceCard
        config={currentConfig}
        onEdit={() =>
          addModal(
            <EditBranchNameModal
              branch={branch}
              currentConfig={currentConfig}
              onSuccess={refresh}
            />
          )
        }
      />

      {hasGitHubSource ? (
        <BranchCISettings
          branch={branch}
          currentConfig={currentConfig}
          orgId={org?.id ?? ''}
          appId={appId}
          onSuccess={refresh}
        />
      ) : null}

      <PreviewConfigSection
        branch={branch}
        currentConfig={currentConfig}
        orgId={org?.id ?? ''}
        appId={appId}
        onSuccess={refresh}
      />

      <Card className="gap-4 p-4 border-red-600/40">
        <div className="flex items-center justify-between gap-2">
          <div className="flex flex-col gap-0.5">
            <Text weight="strong">Delete branch</Text>
            <Text variant="subtext" theme="neutral">
              Permanently deletes this branch, its configs, and its runs.
            </Text>
          </div>
          <Button
            variant="danger"
            onClick={() =>
              addModal(<DeleteBranchModal branch={branch} appId={appId} />)
            }
          >
            <Icon variant="TrashIcon" size={16} />
            Delete branch
          </Button>
        </div>
      </Card>
    </div>
  )
}

export const BranchSettingsPanelContent = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchSettingsCards />
    </BranchProvider>
  )
}

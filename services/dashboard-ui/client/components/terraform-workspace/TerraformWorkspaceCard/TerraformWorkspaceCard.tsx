import type { ReactNode } from 'react'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PulumiState, type TPulumiState } from '@/components/terraform-workspace/PulumiState'
import { TerraformState } from '@/components/terraform-workspace/TerraformState'
import type { TComponentType, TTerraformState } from '@/types'

export interface ITerraformWorkspaceCard {
  currentRevision?: TTerraformState | TPulumiState | null
  actions?: ReactNode
  status?: ReactNode
  componentType?: TComponentType
}

export const TerraformWorkspaceCard = ({
  currentRevision,
  actions,
  status,
  componentType,
}: ITerraformWorkspaceCard) => {
  const isPulumi = componentType === 'pulumi'
  const heading = isPulumi ? 'Pulumi state' : 'Terraform state'

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <Text variant="base" weight="strong">
            {heading}
          </Text>
          {actions && <div className="flex items-center gap-2">{actions}</div>}
        </div>
        {status}
      </div>

      {!currentRevision ? (
        <EmptyState
          variant="diagram"
          emptyTitle="No revisions yet"
          emptyMessage="The workspace has been created, but no state has been written."
        />
      ) : isPulumi ? (
        <PulumiState pulumiState={currentRevision as TPulumiState} />
      ) : (
        <TerraformState terraformState={currentRevision as TTerraformState} />
      )}
    </div>
  )
}

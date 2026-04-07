import type { ReactNode } from 'react'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { TerraformState } from '@/components/terraform-workspace/TerraformState'
import type { TTerraformState } from '@/types'

interface ITerraformWorkspaceCard {
  currentRevision?: TTerraformState | null
  actions?: ReactNode
}

export const TerraformWorkspaceCard = ({
  currentRevision,
  actions,
}: ITerraformWorkspaceCard) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Text variant="base" weight="strong">
          Terraform state
        </Text>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>

      {!currentRevision ? (
        <EmptyState
          variant="diagram"
          emptyTitle="No revisions yet"
          emptyMessage="The workspace has been created, but no state has been written."
        />
      ) : (
        <TerraformState terraformState={currentRevision} />
      )}
    </div>
  )
}

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { BranchesTable } from '@/components/branches/BranchesTable'
import { CreateBranchModal } from '@/components/branches/CreateBranchModal'

export const Branches = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  const { data: branches, isLoading } = useQuery({
    queryKey: ['app-branches', org.id, app.id],
    queryFn: () => getAppBranches({ orgId: org.id!, appId: app.id! }),
    enabled: !!org.id && !!app.id,
  })

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Branches
          </Text>
          <Text variant="subtext" theme="neutral">
            Manage app branches for version control and deployment
          </Text>
        </HeadingGroup>
        <Button onClick={() => setIsCreateModalOpen(true)} variant="primary">
          <Icon variant="Plus" size={16} />
          Create Branch
        </Button>
      </div>

      <BranchesTable branches={branches || []} isLoading={isLoading} />

      <CreateBranchModal
        isVisible={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </div>
  )
}

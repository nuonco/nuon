import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

export interface IBranchRunApprovalItem {
  key: string
  groupName: string
  onReview: () => void
  actions: ReactNode
}

interface IBranchRunApproval {
  items: IBranchRunApprovalItem[]
}

export const BranchRunApproval = ({ items }: IBranchRunApproval) => {
  if (items.length === 0) return null

  return (
    <div className="flex flex-col gap-3">
      {items.map((item) => (
        <Banner key={item.key} className="@container" theme="warn">
          <div className="flex flex-col gap-3">
            <div className="flex flex-col">
              <Text weight="strong">Plan for {item.groupName} requires approval</Text>
              <Text variant="subtext" theme="neutral">
                Review the proposed changes, then approve to deploy to this install group.
              </Text>
            </div>
            <div className="flex flex-wrap items-center justify-end gap-2">
              <Button variant="secondary" onClick={item.onReview}>
                <Icon variant="ListChecksIcon" size={16} />
                Review changes
              </Button>
              {item.actions}
            </div>
          </div>
        </Banner>
      ))}
    </div>
  )
}

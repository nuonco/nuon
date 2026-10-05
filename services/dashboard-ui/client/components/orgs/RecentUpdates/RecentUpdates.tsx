import {
  BranchActivityFeed,
  type TActivityFilter,
  type TBranchActivityItem,
} from '@/components/orgs/BranchActivityFeed'
import { SectionHeader } from '@/components/layout/SectionHeader'

export interface IRecentUpdates {
  items: TBranchActivityItem[]
  isLoading?: boolean
  initialFilter?: TActivityFilter
}

export const RecentUpdates = ({
  items,
  isLoading = false,
  initialFilter,
}: IRecentUpdates) => (
  <>
    <SectionHeader
      title="Recent updates"
      description="Applied and in-flight updates across your app branches, and the installs each one updated."
    />
    <BranchActivityFeed
      items={items}
      isLoading={isLoading}
      initialFilter={initialFilter}
    />
  </>
)

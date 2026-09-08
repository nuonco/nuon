import type { TOrg } from '@/types'
import { OrgProfile } from '../../molecules/OrgProfile'
import { SwitcherMenu } from '../../molecules/SwitcherMenu'

export interface IOrgSwitcherMenu {
  orgs: TOrg[]
  currentOrgId?: string
  search: string
  onSearchChange: (value: string) => void
  onLoadMore: () => void
  isLoading?: boolean
  isLoadingMore?: boolean
  hasMore?: boolean
  hasError?: boolean
}

export const OrgSwitcherMenu = ({
  orgs,
  currentOrgId,
  ...props
}: IOrgSwitcherMenu) => (
  <SwitcherMenu
    {...props}
    items={orgs.flatMap((org) =>
      org?.id && org?.name
        ? [
            {
              id: org.id,
              label: org.name,
              href: `/${org.id}`,
              content: <OrgProfile org={org} avatarSize="sm" />,
            },
          ]
        : []
    )}
    selectedId={currentOrgId}
    loadingContent={<OrgProfile loading avatarSize="sm" />}
    searchLabel="Search organizations"
    searchPlaceholder="Search organizations..."
    emptyTitle="No organizations found"
    errorTitle="Organizations failed to load"
  />
)

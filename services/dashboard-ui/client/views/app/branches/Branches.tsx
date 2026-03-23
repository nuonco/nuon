import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { BranchesTable } from '@/components/branches/BranchesTable'
import { CreateBranchButton } from '@/components/branches/CreateBranchModal'

const CONTAINER_ID = 'app-branches-page'
const LIMIT = 20

export const Branches = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches', org.id, app.id, offset],
    queryFn: () => getAppBranches({ orgId: org.id!, appId: app.id!, limit: LIMIT, offset }),
    enabled: !!org.id && !!app.id,
  })

  const branches = result?.data ?? []
  const pagination = {
    hasNext: result?.pagination?.hasNext ?? false,
    offset,
    limit: LIMIT,
  }

  return (
    <PageSection id={CONTAINER_ID} isScrollable>
      <PageTitle title={`Branches | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
        ]}
      />
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Branches
          </Text>
          <Text variant="subtext" theme="neutral">
            Manage app branches for version control and deployment
          </Text>
        </HeadingGroup>
        <CreateBranchButton />
      </div>
      <BranchesTable branches={branches} isLoading={isLoading} pagination={pagination} />
    </PageSection>
  )
}

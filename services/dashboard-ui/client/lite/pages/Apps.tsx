import { AppsTable } from '../components/organisms/AppsTable'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useOrg } from '../providers/org-provider'

export const Apps = () => {
  const { org, orgId } = useOrg()

  usePageTitle('Apps')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Apps' },
  ])

  return <AppsTable />
}

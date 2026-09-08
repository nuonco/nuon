import { InstallsTable } from '../components/organisms/InstallsTable'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useOrg } from '../providers/org-provider'

export const Installs = () => {
  const { org, orgId } = useOrg()

  usePageTitle('Installs')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Installs' },
  ])

  return <InstallsTable />
}

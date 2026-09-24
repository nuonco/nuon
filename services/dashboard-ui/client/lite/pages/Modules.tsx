import { ModuleManager } from '../components/organisms/ModuleManager'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { useNuonStaff } from '../hooks/use-nuon-staff'
import { usePageTitle } from '../hooks/use-page-title'
import { useOrg } from '../providers/org-provider'
import { NotFound } from './scaffolds'

const ModulesPage = () => {
  const { org, orgId } = useOrg()

  usePageTitle('Modules')
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Modules' },
  ])

  return <ModuleManager />
}

export const Modules = () => {
  const { staff, loading } = useNuonStaff()

  if (loading) return null
  if (!staff) return <NotFound />
  return <ModulesPage />
}

import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { getInstall } from '@/lib'
import { InstallSetup as InstallSetupWizard } from '../components/organisms/InstallSetup'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useOrg } from '../providers/org-provider'

export const InstallSetup = () => {
  const { org, orgId } = useOrg()
  const [params] = useSearchParams()
  const installId = params.get('installId') || undefined
  const { data: install } = useQuery({
    queryKey: ['install', orgId, installId],
    queryFn: () => getInstall({ orgId: orgId!, installId: installId! }),
    enabled: !!orgId && !!installId,
  })

  usePageTitle('Install setup', install?.name)
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    {
      label: 'Installs',
      href: orgId ? `/${orgId}/installs` : undefined,
    },
    ...(installId
      ? [
          {
            label: install?.name,
            href: orgId ? `/${orgId}/installs/${installId}` : undefined,
            loadingWidth: 14,
          },
        ]
      : []),
    { label: 'Install setup' },
  ])

  return <InstallSetupWizard />
}

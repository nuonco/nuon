import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { getInstall } from '@/lib'
import { Card } from '../components/atoms/Card'
import { Text } from '../components/atoms/Text'
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

  return (
    <Card className="min-h-40">
      <Text variant="caption" color="tertiary">
        Page content will be added in a follow-up.
      </Text>
    </Card>
  )
}

import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { getApp } from '@/lib'
import { Card } from '../components/atoms/Card'
import { Text } from '../components/atoms/Text'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useOrg } from '../providers/org-provider'

export const AppSetup = () => {
  const { org, orgId } = useOrg()
  const [params] = useSearchParams()
  const appId = params.get('appId') || undefined
  const { data: app } = useQuery({
    queryKey: ['app', orgId, appId],
    queryFn: () => getApp({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
  })

  usePageTitle('App setup', app?.name)
  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    {
      label: 'Apps',
      href: orgId ? `/${orgId}/apps` : undefined,
    },
    ...(appId
      ? [
          {
            label: app?.name,
            href: orgId ? `/${orgId}/apps/${appId}` : undefined,
            loadingWidth: 14,
          },
        ]
      : []),
    { label: 'App setup' },
  ])

  return (
    <Card className="min-h-40">
      <Text variant="caption" color="tertiary">
        Page content will be added in a follow-up.
      </Text>
    </Card>
  )
}

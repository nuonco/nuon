import { useEffect } from 'react'
import { useNavigate } from 'react-router'

import { Text } from '@/components/common/Text'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { OperationsStudio as OperationsStudioFeature } from '@/components/studio/OperationsStudio'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export function OperationsStudio() {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { app } = useApp()
  const hasRunbookStudio = !!org?.features?.['runbook-studio']

  useEffect(() => {
    if (org && !hasRunbookStudio) {
      navigate(`/${org.id}/apps/${app?.id}/runbooks`, { replace: true })
    }
  }, [org, hasRunbookStudio])

  if (!hasRunbookStudio) return null

  return (
    <PageSection>
      <PageTitle segments={['Runbook studio', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
          {
            path: `/${org?.id}/apps/${app?.id}/studio`,
            text: 'Runbook studio',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="h3" weight="strong">
          Runbook studio
        </Text>
        <Text variant="subtext" theme="neutral">
          Write markdown around executable steps, preview the operator-facing
          document with live install state, then copy the generated files into
          your app config.
        </Text>
      </HeadingGroup>
      <OperationsStudioFeature />
    </PageSection>
  )
}

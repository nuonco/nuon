import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponent, getComponentBuild } from '@/lib'

export const BuildDetail = () => {
  const { componentId, buildId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: component } = useQuery({
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, appId: app.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  const { data: build } = useQuery({
    queryKey: ['component-build', org?.id, componentId, buildId],
    queryFn: () => getComponentBuild({ orgId: org.id, componentId: componentId!, buildId: buildId! }),
    enabled: !!org?.id && !!componentId && !!buildId,
  })

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/components`, text: 'Components' },
          { path: `/${org?.id}/apps/${app?.id}/components/${componentId}`, text: component?.name },
          { path: `/${org?.id}/apps/${app?.id}/components/${componentId}/builds/${buildId}`, text: build?.id },
        ]}
      />
    </PageSection>
  )
}

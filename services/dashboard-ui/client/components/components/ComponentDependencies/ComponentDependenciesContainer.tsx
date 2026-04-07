import { useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'
import { ComponentDependencies } from './ComponentDependencies'

interface IComponentDependenciesContainer {
  deps: string[]
  variant?: 'count' | 'inline'
}

export const ComponentDependenciesContainer = ({
  deps,
  variant = 'count',
}: IComponentDependenciesContainer) => {
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: result, isLoading } = useQuery({
    queryKey: ['components', org?.id, app?.id, 'deps', deps],
    queryFn: () =>
      getComponents({
        orgId: org.id,
        appId: app.id,
        component_ids: deps.toString(),
      }),
    enabled: !!org?.id && !!app?.id && deps.length > 0,
  })

  return (
    <ComponentDependencies
      deps={deps}
      variant={variant}
      components={result?.data ?? []}
      isLoading={isLoading}
      basePath={`/${org.id}/apps/${app.id}/components`}
      pathname={pathname}
    />
  )
}

import { useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'
import { InstallComponentDependencies } from './InstallComponentDependencies'

interface IInstallComponentDependenciesContainer {
  deps: string[]
  variant?: 'count' | 'inline'
  tooltipTitle?: string
}

export const InstallComponentDependenciesContainer = ({
  deps,
  variant = 'count',
  tooltipTitle,
}: IInstallComponentDependenciesContainer) => {
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: result, isLoading } = useQuery({
    queryKey: ['components', org?.id, install?.app_id, 'deps', deps],
    queryFn: () =>
      getComponents({
        orgId: org.id,
        appId: install.app_id,
        component_ids: deps.toString(),
      }),
    enabled: !!org?.id && !!install?.app_id && deps?.length > 0,
  })

  return (
    <InstallComponentDependencies
      deps={deps}
      variant={variant}
      components={result?.data ?? []}
      isLoading={isLoading}
      basePath={`/${org.id}/installs/${install.id}/components`}
      pathname={pathname}
      tooltipTitle={tooltipTitle}
    />
  )
}

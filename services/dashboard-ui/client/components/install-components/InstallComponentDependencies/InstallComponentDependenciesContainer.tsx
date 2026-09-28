import { useLocation } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
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
  const installLink = useInstallLink()

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
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
      basePath={installLink({ installId: install.id, appId: install.app_id, suffix: `/components` })}
      pathname={pathname}
      tooltipTitle={tooltipTitle}
    />
  )
}

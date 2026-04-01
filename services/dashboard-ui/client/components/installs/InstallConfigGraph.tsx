import { lazy, Suspense } from 'react'
import { Skeleton } from '@/components/common/Skeleton'
import { Banner } from '@/components/common/Banner'
import { useInstall } from '@/hooks/use-install'

const ComponentsGraphInline = lazy(() =>
  import('@/components/apps/ConfigGraph/ComponentsGraphRenderer').then(
    (mod) => ({ default: mod.ComponentsGraphInline })
  )
)

export const InstallConfigGraph = () => {
  const { install } = useInstall()

  if (!install.app_id || !install.app_config_id) {
    return (
      <Banner theme="warn">
        Unable to load dependency graph — missing app or config data.
      </Banner>
    )
  }

  return (
    <Suspense fallback={<Skeleton width="100%" height="32rem" />}>
      <ComponentsGraphInline
        appId={install.app_id}
        configId={install.app_config_id}
      />
    </Suspense>
  )
}

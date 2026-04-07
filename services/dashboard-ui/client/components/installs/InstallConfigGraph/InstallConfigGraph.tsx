import { lazy, Suspense } from 'react'
import { Skeleton } from '@/components/common/Skeleton'
import { Banner } from '@/components/common/Banner'

const ComponentsGraphInline = lazy(() =>
  import('@/components/apps/ConfigGraph/ComponentsGraphRenderer').then(
    (mod) => ({ default: mod.ComponentsGraphInline })
  )
)

interface IInstallConfigGraph {
  appId: string | undefined
  appConfigId: string | undefined
}

export const InstallConfigGraph = ({ appId, appConfigId }: IInstallConfigGraph) => {
  if (!appId || !appConfigId) {
    return (
      <Banner theme="warn">
        Unable to load dependency graph — missing app or config data.
      </Banner>
    )
  }

  return (
    <Suspense fallback={<Skeleton width="100%" height="32rem" />}>
      <ComponentsGraphInline appId={appId} configId={appConfigId} />
    </Suspense>
  )
}

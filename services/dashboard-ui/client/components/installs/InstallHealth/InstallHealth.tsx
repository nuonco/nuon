import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'

export interface IInstallHealth {
  resources: ReactNode
  clusterAccessError?: string
}

export const InstallHealth = ({
  resources,
  clusterAccessError,
}: IInstallHealth) => (
  <div className="flex flex-col gap-6">
    {clusterAccessError ? (
      <Banner theme="warn">Cluster access failed: {clusterAccessError}</Banner>
    ) : null}
    {resources}
  </div>
)

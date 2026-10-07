import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TInstallHealthcheck } from '@/types'

export interface IInstallHealth {
  resources: ReactNode
  clusterAccessError?: string
  error?: string
  healthchecks?: TInstallHealthcheck[]
  getHealthcheckHref?: (check: TInstallHealthcheck) => string | undefined
}

export const InstallHealth = ({
  resources,
  clusterAccessError,
  error,
  healthchecks,
  getHealthcheckHref,
}: IInstallHealth) => (
  <div className="flex flex-col gap-6">
    {error ? <Banner theme="error">{error}</Banner> : null}
    {clusterAccessError ? (
      <Banner theme="warn">Cluster access failed: {clusterAccessError}</Banner>
    ) : null}
    {healthchecks?.length ? (
      <Card elevation="1">
        <Text weight="strong">Health checks</Text>
        {healthchecks.map((check) => {
          const href = getHealthcheckHref?.(check)
          return (
            <div
              key={check.action_id}
              className="flex items-center justify-between gap-3"
            >
              {href ? (
                <Link href={href}>{check.name}</Link>
              ) : (
                <Text>{check.name}</Text>
              )}
              <Status variant="badge" status={check.status} />
            </div>
          )
        })}
      </Card>
    ) : null}
    {resources}
  </div>
)

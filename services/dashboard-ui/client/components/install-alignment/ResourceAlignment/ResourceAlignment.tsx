import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TAPIError, TInstallAlignmentResource } from '@/types'
import { hasOngoingDeployment, isDrifted } from '../alignment-utils'
import { OngoingDeployments } from '../DriftedResources/OngoingDeployments'

export interface IResourceAlignment {
  card?: ReactNode
  error?: TAPIError | null
  label: string
  loading?: boolean
  onRetry?: () => void
  resource?: TInstallAlignmentResource
}

export const ResourceAlignment = ({
  card,
  error,
  label,
  loading = false,
  onRetry,
  resource,
}: IResourceAlignment) => {
  if (error && !resource) {
    return (
      <Banner theme="error" data-testid="alignment-error">
        <div className="flex flex-wrap items-center justify-between gap-3 w-full">
          <Text variant="subtext">
            Unable to load drift for the {label}.{' '}
            {error.description || error.error}
          </Text>
          {onRetry ? (
            <Button size="sm" onClick={onRetry}>
              Try again
            </Button>
          ) : null}
        </div>
      </Banner>
    )
  }
  if (loading && !resource) {
    return (
      <div data-testid="alignment-loading" aria-busy>
        <Skeleton height="5rem" width="100%" />
      </div>
    )
  }
  if (!resource) return null
  if (hasOngoingDeployment(resource)) {
    return (
      <OngoingDeployments
        resources={[resource]}
        renderCard={() => card}
        view="list"
        pageSize={1}
      />
    )
  }
  if (isDrifted(resource)) {
    return (
      <section aria-label="Drifted" data-testid="drifted-section">
        {card}
      </section>
    )
  }

  const inSync = resource.category === 'in-sync'
  const message = inSync
    ? `The ${label} matches the selected commit. No changes are needed.`
    : resource.category === 'disabled'
      ? `The ${label} is disabled, so it is not compared with the selected commit.`
      : `Drift could not be determined for the ${label}.`

  return (
    <div
      className="flex flex-wrap items-center gap-2"
      data-testid="alignment-summary"
      data-category={resource.category}
    >
      <Icon
        variant={inSync ? 'CheckCircleIcon' : 'InfoIcon'}
        size={18}
        className={
          inSync
            ? 'text-green-800 dark:text-green-500'
            : 'text-cool-grey-600 dark:text-cool-grey-500'
        }
      />
      <Text variant="subtext">{message}</Text>
    </div>
  )
}

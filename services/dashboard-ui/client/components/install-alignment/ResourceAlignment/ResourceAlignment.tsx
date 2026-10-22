import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
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
          {onRetry ? <Button onClick={onRetry}>Try again</Button> : null}
        </div>
      </Banner>
    )
  }
  if (loading && !resource) {
    return (
      <div
        className="flex items-center gap-2"
        role="status"
        data-testid="alignment-loading"
        aria-busy
      >
        <Loading />
        <Text variant="subtext" theme="neutral">
          Loading {label} alignment
        </Text>
      </div>
    )
  }
  if (!resource) return null

  const inSync = resource.category === 'in-sync'
  const message = inSync
    ? `The ${label} matches the selected commit. No changes are needed.`
    : resource.category === 'disabled'
      ? `The ${label} is disabled, so it is not compared with the selected commit.`
      : `Drift could not be determined for the ${label}.`

  return (
    <div className="flex flex-col gap-3">
      {hasOngoingDeployment(resource) ? (
        <OngoingDeployments
          resources={[resource]}
          renderCard={() => card}
          view="list"
          pageSize={1}
        />
      ) : isDrifted(resource) ? (
        <section aria-label="Drifted" data-testid="drifted-section">
          {card}
        </section>
      ) : (
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
      )}
      {error ? (
        <div role="alert">
          <Text variant="subtext" theme="error">
            Refreshing resources failed. Showing the last loaded results.{' '}
            {error.description || error.error}
          </Text>
        </div>
      ) : null}
    </div>
  )
}

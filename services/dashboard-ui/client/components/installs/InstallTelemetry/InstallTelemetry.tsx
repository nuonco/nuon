import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Toggle } from '@/components/common/form/Toggle'
import type { TAPIError } from '@/types'

export interface IInstallTelemetry {
  enabled: boolean
  isLoading?: boolean
  isPending?: boolean
  hasEndpoint?: boolean
  isRunnerActive?: boolean
  isInherited?: boolean
  isManagedByConfig?: boolean
  canUseOrgDefault?: boolean
  onUseOrgDefault?: () => void
  error?: TAPIError | null
  onToggle: (enabled: boolean) => void
  onRetry: () => void
}

export const InstallTelemetry = ({
  enabled,
  isLoading = false,
  isPending = false,
  hasEndpoint = true,
  isRunnerActive = true,
  isInherited,
  isManagedByConfig = false,
  canUseOrgDefault = true,
  onUseOrgDefault,
  error,
  onToggle,
  onRetry,
}: IInstallTelemetry) => {
  return (
    <div className="flex flex-col gap-3">
      {(!hasEndpoint || !isRunnerActive) && !isLoading && !error ? (
        <Text variant="subtext" theme="warn">
          {!hasEndpoint &&
            'Update the install stack to add the missing private telemetry endpoint. '}
          {!isRunnerActive &&
            'Telemetry might not be flowing because the runner is not active. '}
          {enabled && 'Forwarding can still be disabled.'}
        </Text>
      ) : null}
      {isLoading ? (
        <Text variant="subtext" theme="neutral" aria-live="polite">
          Loading settings...
        </Text>
      ) : error ? (
        <>
          <Banner theme="error" role="alert">
            {error.description ||
              error.error ||
              'Unable to load telemetry settings.'}
          </Banner>
          <Button variant="secondary" className="w-fit" onClick={onRetry}>
            Retry settings
          </Button>
        </>
      ) : (
        <Tooltip
          tipContent={
            isManagedByConfig
              ? 'Managed by config. Disable config sync to edit.'
              : 'Cannot enable telemetry. Telemetry endpoint missing from the install stack. Ensure your stack enables telemetry ingress.'
          }
          disableHover={!isManagedByConfig && (enabled || hasEndpoint)}
          tabIndex={
            isManagedByConfig || (!enabled && !hasEndpoint) ? 0 : undefined
          }
          tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs"
          className="rounded-md focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80"
        >
          <Toggle
            className="w-fit rounded-md focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80"
            checked={enabled}
            disabled={
              isPending || isManagedByConfig || (!enabled && !hasEndpoint)
            }
            label={isPending ? 'Saving...' : 'Enable telemetry'}
            aria-label="Enable telemetry"
            onChange={onToggle}
          />
        </Tooltip>
      )}
      {!isLoading && !error && isInherited !== undefined ? (
        isInherited ? (
          <Text variant="subtext" theme="neutral">
            Using org default
          </Text>
        ) : (
          <Button
            variant="secondary"
            className="w-fit"
            disabled={isPending || isManagedByConfig || !canUseOrgDefault}
            tooltipProps={
              isManagedByConfig
                ? {
                    tipContent:
                      'Managed by config. Disable config sync to edit.',
                  }
                : !canUseOrgDefault
                  ? {
                      tipContent:
                        'A telemetry endpoint is required. Ensure your install stack enables telemetry ingress.',
                    }
                  : undefined
            }
            onClick={onUseOrgDefault}
          >
            Use org default
          </Button>
        )
      ) : null}
      <Link href="https://docs.nuon.co/guides/byoc/telemetry" isExternal>
        View telemetry setup
      </Link>
    </div>
  )
}

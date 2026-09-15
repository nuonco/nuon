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
  hasSetup?: boolean
  error?: TAPIError | null
  onToggle: (enabled: boolean) => void
  onRetry: () => void
}

export const InstallTelemetry = ({
  enabled,
  isLoading = false,
  isPending = false,
  hasSetup = true,
  error,
  onToggle,
  onRetry,
}: IInstallTelemetry) => (
  <div className="flex flex-col gap-3">
    {!hasSetup && !isLoading && !error ? (
      <Text variant="subtext" theme="warn">
        {enabled
          ? 'Install telemetry setup is unavailable. Forwarding can still be disabled.'
          : 'A runner and private telemetry endpoint are required. Update the install stack to enable telemetry.'}
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
        tipContent="Cannot enable telemetry — a runner and private endpoint are required"
        disableHover={enabled || hasSetup}
        tabIndex={!enabled && !hasSetup ? 0 : undefined}
        className="rounded-md focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80"
      >
        <Toggle
          className="w-fit rounded-md focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-2 focus-visible:outline-primary-400/80"
          checked={enabled}
          disabled={isPending || (!enabled && !hasSetup)}
          label={isPending ? 'Saving...' : 'Enable telemetry'}
          aria-label="Enable telemetry"
          onChange={onToggle}
        />
      </Tooltip>
    )}
    <Link href="https://docs.nuon.co/guides/byoc/telemetry" isExternal>
      View telemetry setup
    </Link>
  </div>
)

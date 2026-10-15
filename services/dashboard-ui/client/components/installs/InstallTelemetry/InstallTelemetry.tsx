import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import type { TAPIError, TInstallTelemetrySettings } from '@/types'

export interface IInstallTelemetry {
  settings?: TInstallTelemetrySettings
  isLoading?: boolean
  error?: TAPIError | null
  onManage: () => void
  onRetry: () => void
}

export const InstallTelemetry = ({
  settings,
  isLoading,
  error,
  onManage,
  onRetry,
}: IInstallTelemetry) => (
  <div className="flex flex-col gap-3">
    <LabeledValue label="Forwarding" loading={isLoading}>
      {settings
        ? `${settings.enabled ? 'On' : 'Off'}${settings.override == null ? ' · Org default' : ' · Install override'}`
        : 'Unknown'}
    </LabeledValue>
    {!error && settings && !settings.relay_configured ? (
      <Text variant="subtext" theme="warn">
        Relay setup required
      </Text>
    ) : null}
    {error ? (
      <Banner theme="error">
        {error.description || error.error || 'Telemetry settings load failed'}
      </Banner>
    ) : null}
    <Button
      variant="secondary"
      className="w-fit"
      onClick={error ? onRetry : onManage}
    >
      {error ? 'Retry settings' : 'Manage telemetry'}
    </Button>
  </div>
)

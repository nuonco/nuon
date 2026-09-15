import { useState } from 'react'
import { InstallTelemetry, type IInstallTelemetry } from './InstallTelemetry'

export default { title: 'Installs/InstallTelemetry' }

const Example = (props: Partial<IInstallTelemetry>) => {
  const [enabled, setEnabled] = useState(props.enabled ?? false)
  return (
    <div className="max-w-sm">
      <InstallTelemetry
        onRetry={() => {}}
        {...props}
        enabled={enabled}
        onToggle={setEnabled}
      />
    </div>
  )
}

export const Disabled = () => <Example />
export const Enabled = () => <Example enabled />
export const Saving = () => <Example isPending />
export const Loading = () => <Example isLoading />
export const SetupRequired = () => <Example hasSetup={false} />
export const SetupUnavailable = () => <Example enabled hasSetup={false} />
export const LoadError = () => (
  <Example
    error={{
      error: 'Request failed',
      description: 'Unable to load telemetry settings.',
      user_error: false,
    }}
  />
)

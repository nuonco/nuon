import { useState } from 'react'
import { InstallTelemetry, type IInstallTelemetry } from './InstallTelemetry'

export default { title: 'Installs/InstallTelemetry' }

const Example = (props: Partial<IInstallTelemetry>) => {
  const [enabled, setEnabled] = useState(props.enabled ?? false)
  return (
    <div className="max-w-sm p-6">
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
export const Inherited = () => <Example enabled isInherited />
export const Override = () => (
  <Example isInherited={false} onUseOrgDefault={() => {}} />
)
export const ConfigManaged = () => (
  <Example enabled isInherited={false} isManagedByConfig />
)
export const ResetEndpointPending = () => (
  <Example enabled isInherited={false} canUseOrgDefault={false} />
)
export const Saving = () => <Example isPending />
export const Loading = () => <Example isLoading />
export const EndpointMissing = () => <Example hasEndpoint={false} />
export const RunnerOffline = () => <Example isRunnerActive={false} />
export const SetupRequired = () => (
  <Example hasEndpoint={false} isRunnerActive={false} />
)
export const SetupUnavailable = () => (
  <Example enabled hasEndpoint={false} isRunnerActive={false} />
)
export const LoadError = () => (
  <Example
    error={{
      error: 'Request failed',
      description: 'Unable to load telemetry settings.',
      user_error: false,
    }}
  />
)

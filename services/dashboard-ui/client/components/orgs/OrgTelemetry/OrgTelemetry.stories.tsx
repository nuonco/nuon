import { ModalStory } from '@/components/__stories__/helpers'
import type { TAPIError } from '@/types'
import { OrgTelemetryModal } from './OrgTelemetry'

export default { title: 'Orgs/OrgTelemetry' }

const Example = ({
  enabled = false,
  isPending = false,
  error = null,
}: {
  enabled?: boolean
  isPending?: boolean
  error?: TAPIError | null
}) => (
  <ModalStory>
    <OrgTelemetryModal
      orgName="acme"
      enabled={enabled}
      isPending={isPending}
      error={error}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const Disabled = () => <Example />
export const Enabled = () => <Example enabled />
export const Saving = () => <Example enabled isPending />
export const SaveError = () => (
  <Example
    error={{
      error: 'Only org admins can change the telemetry default',
      description: '',
      user_error: true,
    }}
  />
)

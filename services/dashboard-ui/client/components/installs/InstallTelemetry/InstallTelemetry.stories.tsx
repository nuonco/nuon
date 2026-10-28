import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ModalStory } from '@/components/__stories__/helpers'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { InstallTelemetry } from './InstallTelemetry'
import { InstallTelemetryModal } from './InstallTelemetryModal'
import { CollectorSetupContainer } from './CollectorSetupContainer'

export default { title: 'Features / Installs / Telemetry' }

const orgId = 'org-acme'
const install = {
  id: 'install-acme',
  name: 'acme-production',
  app_id: 'app-acme',
  runner_id: 'runner-acme',
  runner_status: 'offline',
}

const SetupModal = ({
  managed = false,
  setupRequired = false,
  member = false,
  runner = false,
  ...props
}: IModal & {
  managed?: boolean
  setupRequired?: boolean
  member?: boolean
  runner?: boolean
}) => {
  const [saved, setSaved] = useState(false)
  const [client] = useState(() => {
    const client = new QueryClient({
      defaultOptions: { queries: { staleTime: Infinity, retry: false } },
    })
    client.setQueryData(['account'], {
      roles: [
        { org_id: orgId, role_type: member ? 'org_member' : 'org_admin' },
      ],
    })
    client.setQueryData(
      ['install-telemetry-collector', orgId, install.id],
      setupRequired || runner ? null : { account_id: 'account-collector' }
    )
    client.setQueryData(
      ['install-components', orgId, install.id, 'telemetry-collector'],
      {
        data: runner
          ? []
          : [{ component_id: 'component-collector', enabled: !setupRequired }],
        pagination: { hasNext: false },
      }
    )
    client.setQueryData(
      ['tokens', orgId, 'account-collector'],
      [
        {
          id: 'token-bootstrap-old',
          account_id: 'account-collector',
          expires_at: '2025-01-01T00:00:00Z',
        },
      ]
    )
    client.setQueryData(['install-stack', orgId, install.id], {
      install_stack_outputs: {
        data_contents: {
          telemetry_endpoint: runner
            ? 'http://internal-acme-telemetry-1234567890abcdef.us-east-1.elb.amazonaws.com:4318'
            : '',
        },
      },
    })
    return client
  })
  return (
    <QueryClientProvider client={client}>
      <InstallTelemetryModal
        installName={install.name}
        settings={{
          enabled: false,
          override: false,
          org_default: true,
          relay_configured: !setupRequired,
        }}
        isManagedByConfig={managed}
        isPending={false}
        error={null}
        {...props}
        onSubmit={() => setSaved(true)}
      >
        {saved ? (
          <div role="status">
            <Text>Telemetry settings updated</Text>
          </div>
        ) : null}
        <CollectorSetupContainer
          install={{ ...install, runner_status: runner ? 'active' : 'offline' }}
          orgId={orgId}
          onNavigate={() => {}}
        />
      </InstallTelemetryModal>
    </QueryClientProvider>
  )
}

export const Default = () => (
  <ModalStory>
    <SetupModal />
  </ModalStory>
)
export const SetupRequired = () => (
  <ModalStory>
    <SetupModal setupRequired />
  </ModalStory>
)
export const ConfigManaged = () => (
  <ModalStory>
    <SetupModal managed />
  </ModalStory>
)
export const Member = () => (
  <ModalStory>
    <SetupModal member />
  </ModalStory>
)
export const RunnerCollector = () => (
  <ModalStory>
    <SetupModal runner />
  </ModalStory>
)
export const Summary = () => (
  <div className="max-w-sm p-6">
    <InstallTelemetry
      settings={{
        enabled: true,
        override: null,
        org_default: true,
        relay_configured: true,
      }}
      onManage={() => {}}
      onRetry={() => {}}
    />
  </div>
)
export const Loading = () => (
  <InstallTelemetry isLoading onManage={() => {}} onRetry={() => {}} />
)
export const LoadError = () => (
  <InstallTelemetry
    error={{
      error: 'Telemetry settings load failed',
      description: 'Retry loading telemetry settings.',
      user_error: false,
    }}
    onManage={() => {}}
    onRetry={() => {}}
  />
)

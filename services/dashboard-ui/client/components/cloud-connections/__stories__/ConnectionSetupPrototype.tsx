import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Text } from '@/components/common/Text'
import type { TCloudConnection } from '@/types'
import { ConnectionDetail } from '../ConnectionDetail/ConnectionDetail'
import { ConnectionForm } from '../ConnectionWizard/ConnectionForm'
import { ConnectionWizard } from '../ConnectionWizard/ConnectionWizard'
import { RunInCloud } from '../ConnectionWizard/RunInCloud'
import { RepositoryAccessDemo } from './RepositoryAccessDemo'
import {
  PageStory,
  verifiedECRConnection,
  verifiedGARConnection,
  verifiedACRConnection,
  noop,
} from './fixtures'

export const prototypeConnections = [
  { ...verifiedECRConnection, name: 'acme-ecr-images' },
  { ...verifiedGARConnection, name: 'acme-gar-images' },
  { ...verifiedACRConnection, name: 'acme-acr-images' },
]

export const ConnectionSetupPrototype = ({
  connection,
  existing = false,
  preview = false,
  state = 'ready',
}: {
  connection: TCloudConnection
  existing?: boolean
  preview?: boolean
  state?: 'ready' | 'loading' | 'generating' | 'error'
}) => {
  const [created, setCreated] = useState<TCloudConnection>()
  const picker = useForm({ defaultValues: { connection: connection.id } })
  const selectedID = useStore(picker.store, (s) => s.values.connection)
  const selected =
    prototypeConnections.find((item) => item.id === selectedID) || connection
  const saved = existing ? selected : created
  const repositoryAccess =
    saved && state !== 'loading' ? (
      <RepositoryAccessDemo
        key={saved.id}
        connection={saved}
        generated={preview}
        state={state}
      />
    ) : undefined
  return (
    <PageStory>
      <Text variant="subtext" theme="neutral">
        Interactive story · Sample data and output. Nothing is saved or run in
        the cloud.
      </Text>
      {existing && (
        <div className="max-w-2xl">
          <picker.Field name="connection">
            {(field) => (
              <FormSelect
                field={field}
                id="story-connection"
                labelProps={{ labelText: 'Existing connection' }}
                disabled={state === 'loading' || state === 'generating'}
                options={prototypeConnections.map((item) => ({
                  value: item.id,
                  label: item.name,
                  content: <Text family="mono">{item.name}</Text>,
                }))}
              />
            )}
          </picker.Field>
        </div>
      )}
      {existing ? (
        <ConnectionDetail
          connection={state === 'loading' ? undefined : selected}
          basePath="/org-mock-001/settings/cloud-connections/example"
          tab="overview"
          onVerify={noop}
          onDelete={noop}
          repositoryAccess={repositoryAccess}
        />
      ) : (
        <ConnectionWizard connection={created}>
          {created ? (
            <>
              <RunInCloud
                connection={created}
                setup={created.setup}
                hasInitialSetup
              />
              {repositoryAccess}
            </>
          ) : (
            <ConnectionForm
              values={{
                name: connection.name,
                platform: connection.platform,
                target_id: connection.target_id,
                preset: connection.preset,
                role_name: connection.principal.split('/').at(-1),
                service_account_name: connection.principal.split('@')[0],
                gcp_project_number: connection.gcp_project_number,
                azure_tenant_id: connection.azure_tenant_id,
                principal: connection.principal,
                ecr_region: connection.default_region,
                gar_location: connection.default_region,
                azure_resource_group:
                  connection.azure_registry_id?.split('/')[4],
                azure_registry_name: connection.azure_registry_id
                  ?.split('/')
                  .at(-1),
                azure_registry_login_server:
                  connection.azure_registry_login_server,
              }}
              onCreate={(body) =>
                setCreated({ ...connection, ...body, status: 'pending' })
              }
            />
          )}
        </ConnectionWizard>
      )}
    </PageStory>
  )
}

import { useState } from 'react'
import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { createCloudConnection, getCloudConnectionSetup } from '@/lib'
import type {
  TCloudConnectionSetup,
  TCreateCloudConnectionRequest,
} from '@/types'
import { ConnectionSetupMode } from '../ConnectionSetupMode'
import { PrincipalSharing } from '../PrincipalSharing'
import {
  sharedPrincipalConnections,
  useCloudConnection,
  useCloudConnections,
  useVerifyCloudConnection,
} from '../queries'
import { isImagePreset } from '../repository-scope'
import { ConnectionForm } from './ConnectionForm'
import { ConnectionWizard } from './ConnectionWizard'
import { RunInCloud } from './RunInCloud'
import { VerifyConnection } from './VerifyConnection'

export const ConnectionWizardContainer = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId?: string
}) => {
  const [params, setParams] = useSearchParams()
  const location = useLocation()
  const setupState = location.state as {
    connectionId?: string
    initialSetup?: TCloudConnectionSetup
  } | null
  const initialSetup =
    setupState?.connectionId === connectionId
      ? setupState?.initialSetup
      : undefined
  const navigate = useNavigate()
  const client = useQueryClient()
  const connection = useCloudConnection(orgId, connectionId || '')
  const identities = useCloudConnections(orgId)
  const shared = connection.data
    ? sharedPrincipalConnections(identities.data || [], connection.data)
    : []
  const [setupMode, setSetupMode] = useState<boolean>()
  const existingPrincipal =
    setupMode ?? (shared.length > 0 || connection.data?.status === 'verified')
  const setup = useQuery({
    queryKey: [
      'cloud-connection-setup',
      orgId,
      connectionId,
      existingPrincipal,
    ],
    queryFn: () =>
      getCloudConnectionSetup({
        orgId,
        connectionId: connectionId!,
        existingPrincipal,
      }),
    enabled: !!connectionId && !!connection.data && existingPrincipal,
  })
  const verify = useVerifyCloudConnection(orgId, connectionId || '')
  const setStep = (next: string) =>
    setParams({ step: next }, { state: location.state })
  const create = useMutation({
    mutationFn: (body: TCreateCloudConnectionRequest) =>
      createCloudConnection({ orgId, body }),
    onSuccess: (created) => {
      client.setQueryData(['cloud-connections', orgId, created.id], created)
      client.invalidateQueries({ queryKey: ['cloud-connections', orgId] })
      navigate(`/${orgId}/settings/cloud-connections/${created.id}/setup`, {
        replace: true,
        state: { connectionId: created.id, initialSetup: created.setup },
      })
    },
  })
  const detailHref = `/${orgId}/settings/cloud-connections/${connectionId}`
  return (
    <ConnectionWizard
      connection={connectionId ? connection.data : undefined}
      created={!!connectionId}
      loading={!!connectionId && connection.isLoading}
    >
      {!connectionId ? (
        <ConnectionForm
          orgId={orgId}
          connections={identities.data}
          onCreate={(body) => create.mutate(body)}
          isPending={create.isPending}
          error={create.error}
        />
      ) : (
        <>
          <FormErrorBanner
            error={connection.error || identities.error || setup.error}
            fallback="Connection loading failed"
          />
          {connection.data &&
            (params.get('step') === 'verify' ? (
              <VerifyConnection
                connection={connection.data}
                isVerifying={verify.isPending || connection.isVerifying}
                verificationTimedOut={connection.verificationTimedOut}
                error={verify.error}
                onVerify={() => verify.mutate()}
                setupHref={`${detailHref}/setup`}
                detailHref={detailHref}
              />
            ) : (
              <>
                <RunInCloud
                  connection={connection.data}
                  orgId={orgId}
                  setup={
                    identities.isPending
                      ? undefined
                      : existingPrincipal
                        ? setup.data
                        : initialSetup || connection.data.setup
                  }
                  hasInitialSetup={!existingPrincipal && !!initialSetup}
                  setupMode={
                    <>
                      <PrincipalSharing connections={shared} orgId={orgId} />
                      <ConnectionSetupMode
                        platform={connection.data.platform}
                        existingPrincipal={existingPrincipal}
                        onChange={setSetupMode}
                      />
                    </>
                  }
                />
                <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
                  <Button href={detailHref}>View connection</Button>
                  <Button onClick={() => setStep('verify')}>
                    Review verification
                  </Button>
                  {isImagePreset(connection.data.preset) && (
                    <Button href={`/${orgId}/settings/image-repositories/add`}>
                      Add image repository
                    </Button>
                  )}
                </div>
              </>
            ))}
        </>
      )}
    </ConnectionWizard>
  )
}

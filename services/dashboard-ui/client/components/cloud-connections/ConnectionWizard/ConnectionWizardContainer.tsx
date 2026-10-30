import { useLocation, useNavigate, useSearchParams } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { createCloudConnection } from '@/lib'
import type {
  TCloudConnectionSetup,
  TCreateCloudConnectionRequest,
} from '@/types'
import { AddRepositoryAccess } from '../AddRepositoryAccess'
import { useCloudConnection, useVerifyCloudConnection } from '../queries'
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
  const verify = useVerifyCloudConnection(orgId, connectionId || '')
  const setStep = (next: number) =>
    setParams({ step: String(next) }, { state: location.state })
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
          onCreate={(body) => create.mutate(body)}
          isPending={create.isPending}
          error={create.error}
        />
      ) : (
        <>
          <FormErrorBanner
            error={connection.error}
            fallback="Connection loading failed"
          />
          {connection.data &&
            (params.get('step') === '4' ? (
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
                  setup={initialSetup || connection.data.setup}
                  hasInitialSetup={!!initialSetup}
                />
                <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
                  <Button href={detailHref}>View connection</Button>
                  <Button onClick={() => setStep(4)}>
                    Review verification
                  </Button>
                </div>
                {['ecr', 'gar', 'acr'].includes(connection.data.preset) && (
                  <AddRepositoryAccess
                    key={connectionId}
                    connection={connection.data}
                    orgId={orgId}
                  />
                )}
              </>
            ))}
        </>
      )}
    </ConnectionWizard>
  )
}

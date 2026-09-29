import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Loading } from '@/components/common/Loading'
import { createCloudConnection } from '@/lib'
import type { TCloudConnection } from '@/types'
import { useCloudConnection, useVerifyCloudConnection } from '../queries'
import { AccountForm } from './AccountForm'
import { AccessForm } from './AccessForm'
import { ConnectionWizard } from './ConnectionWizard'
import { RunInCloud } from './RunInCloud'
import { VerifyConnection } from './VerifyConnection'
import type { TAccountValues } from './schema'

export const ConnectionWizardContainer = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId?: string
}) => {
  const [values, setValues] = useState<TAccountValues>()
  const [params, setParams] = useSearchParams()
  const navigate = useNavigate()
  const client = useQueryClient()
  const connection = useCloudConnection(orgId, connectionId || '')
  const verify = useVerifyCloudConnection(orgId, connectionId || '')
  const step = connectionId
    ? params.get('step') === '4'
      ? 4
      : 3
    : values && params.get('step') === '2'
      ? 2
      : 1
  const setStep = (next: number) => setParams({ step: String(next) })
  const create = useMutation({
    mutationFn: (preset: TCloudConnection['preset']) =>
      createCloudConnection({
        orgId,
        body: {
          name: values!.name,
          target_id: values!.target_id,
          principal: `arn:aws:iam::${values!.target_id}:role/${values!.role_name}`,
          platform: 'aws',
          preset,
        },
      }),
    onSuccess: (created) => {
      client.setQueryData(['cloud-connections', orgId, created.id], created)
      client.invalidateQueries({ queryKey: ['cloud-connections', orgId] })
      navigate(`/${orgId}/cloud-connections/${created.id}/setup`, {
        replace: true,
      })
    },
  })
  const detailHref = `/${orgId}/cloud-connections/${connectionId}`
  return (
    <ConnectionWizard step={step} created={!!connectionId} onStep={setStep}>
      {step === 1 && (
        <AccountForm
          values={values}
          onContinue={(next) => {
            setValues(next)
            setStep(2)
          }}
        />
      )}
      {step === 2 && (
        <AccessForm
          onBack={() => setStep(1)}
          onContinue={(preset) => create.mutate(preset)}
          isPending={create.isPending}
          error={create.error}
        />
      )}
      {connectionId && (
        <FormErrorBanner
          error={connection.error}
          fallback="Cloud connection failed to load"
        />
      )}
      {connectionId && connection.isLoading && <Loading />}
      {step === 3 && connection.data && (
        <div className="flex flex-col gap-6">
          <RunInCloud
            connection={connection.data}
            setup={connection.data.setup}
          />
          <div className="flex justify-end gap-2 border-t pt-4">
            <Button variant="secondary" href={detailHref}>
              View connection
            </Button>
            <Button variant="primary" onClick={() => setStep(4)}>
              Review verification
            </Button>
          </div>
        </div>
      )}
      {step === 4 && connection.data && (
        <VerifyConnection
          connection={connection.data}
          isVerifying={verify.isPending || connection.isVerifying}
          verificationTimedOut={connection.verificationTimedOut}
          error={verify.error}
          onVerify={() => verify.mutate()}
          setupHref={`${detailHref}/setup`}
          detailHref={detailHref}
        />
      )}
    </ConnectionWizard>
  )
}

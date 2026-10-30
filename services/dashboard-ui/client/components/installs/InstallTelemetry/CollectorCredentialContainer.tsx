import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createToken } from '@/lib/ctl-api/tokens/create-token'
import { ensureInstallTelemetryCollector } from '@/lib/ctl-api/installs/install-telemetry-collector'
import {
  CollectorCredentialModal,
  type TCollectorCredentialValues,
} from './CollectorCredential'

export const CollectorCredentialModalContainer = ({
  installId,
  orgId,
  ...props
}: {
  installId: string
  orgId: string
} & Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const [credential, setCredential] = useState<
    (TCollectorCredentialValues & { token: string }) | null
  >(null)
  const mutation = useMutation({
    retry: false,
    mutationFn: async (values: TCollectorCredentialValues) => {
      const collector = await ensureInstallTelemetryCollector({
        installId,
        orgId,
      })
      await queryClient.invalidateQueries({
        queryKey: ['install-telemetry-collector', orgId, installId],
      })
      if (!collector.account_id)
        throw { error: 'Collector account was not returned', user_error: false }
      const result = await createToken({
        orgId,
        body: {
          account_id: collector.account_id,
          name: `telemetry-${installId}`,
          duration: values.duration,
          invalidate: false,
        },
      })
      if (!result?.token)
        throw {
          error:
            'Credential was not returned. Check issued tokens before trying again.',
          user_error: false,
        }
      setCredential({ ...values, token: result.token })
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens', orgId] })
      queryClient.invalidateQueries({ queryKey: ['service-accounts', orgId] })
    },
  })
  return (
    <CollectorCredentialModal
      installId={installId}
      credential={credential}
      isPending={mutation.isPending}
      error={mutation.error}
      onSubmit={(values) => mutation.mutate(values)}
      onDone={() => {
        setCredential(null)
        removeModal(props.modalId)
      }}
      {...props}
    />
  )
}

import { useState } from 'react'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  getTriggerIngressURL,
  getTriggerEventTypes,
  revealTriggerSecret,
  revokeTriggerSecret,
  rotateTriggerIngressURL,
  rotateTriggerSecret,
} from '@/lib'
import type { TAPIError, TTrigger, TRotateTriggerSecretResponse } from '@/types'
import { TriggerOverview, type TRevealedSecret } from './TriggerOverview'

type TTriggerAction = {
  triggerId: string
  orgId: string
  triggerName?: string
}

const RotateIngressURLModal = ({
  triggerId,
  orgId,
  triggerName,
  ...props
}: TTriggerAction & IModal) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const mutation = useMutation({
    mutationFn: () => rotateTriggerIngressURL({ triggerId, orgId }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['trigger', orgId, triggerId],
      })
      queryClient.invalidateQueries({
        queryKey: ['triggers', orgId],
      })
      queryClient.invalidateQueries({
        queryKey: ['event-trigger-ingress-url', orgId, triggerId],
      })
      removeModal(props.modalId)
    },
  })
  return (
    <Modal
      heading="Replace ingress URL?"
      primaryActionTrigger={{
        children: mutation.isPending ? 'Replacing...' : 'Replace ingress URL',
        disabled: mutation.isPending,
        onClick: () => mutation.mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      {mutation.error ? (
        <Banner theme="error">
          {(mutation.error as TAPIError)?.error ||
            'Unable to replace the ingress URL.'}
        </Banner>
      ) : null}
      <Text>
        This replaces the ingress URL for {triggerName || 'this trigger'}.
        Requests to the previous URL will stop working immediately.
      </Text>
    </Modal>
  )
}

const RotateSecretModal = ({
  triggerId,
  orgId,
  triggerName,
  ...props
}: TTriggerAction & IModal) => {
  const queryClient = useQueryClient()
  const [result, setResult] = useState<TRotateTriggerSecretResponse>()
  const mutation = useMutation({
    mutationFn: () => rotateTriggerSecret({ triggerId, orgId }),
    onSuccess: (response) => {
      setResult(response)
      queryClient.invalidateQueries({
        queryKey: ['trigger', orgId, triggerId],
      })
      queryClient.invalidateQueries({
        queryKey: ['triggers', orgId],
      })
    },
  })

  if (result)
    return (
      <Modal heading="Secret rotated" {...props}>
        <Banner theme="warn">
          Copy this secret now. It cannot be retrieved after this dialog closes.
        </Banner>
        <div>
          <Text variant="subtext" theme="neutral">
            Secret
          </Text>
          <ClickToCopy>
            <Code>{result?.secret || '—'}</Code>
          </ClickToCopy>
        </div>
      </Modal>
    )

  return (
    <Modal
      heading="Rotate secret?"
      primaryActionTrigger={{
        children: mutation.isPending ? 'Rotating...' : 'Rotate secret',
        disabled: mutation.isPending,
        onClick: () => mutation.mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      {mutation.error ? (
        <Banner theme="error">
          {(mutation.error as TAPIError)?.error ||
            'Unable to rotate the secret.'}
        </Banner>
      ) : null}
      <Text>
        A new secret will be created for {triggerName || 'this trigger'}. The
        current secret remains valid for 24 hours so you can update the sender
        without downtime.
      </Text>
    </Modal>
  )
}

const RevokeSecretModal = ({
  triggerId,
  orgId,
  secretId,
  ...props
}: TTriggerAction & { secretId: string } & IModal) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const mutation = useMutation({
    mutationFn: () => revokeTriggerSecret({ triggerId, orgId, secretId }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['trigger', orgId, triggerId],
      })
      queryClient.invalidateQueries({
        queryKey: ['triggers', orgId],
      })
      removeModal(props.modalId)
    },
  })
  return (
    <Modal
      heading="Revoke secret?"
      primaryActionTrigger={{
        children: mutation.isPending ? 'Revoking...' : 'Revoke secret',
        disabled: mutation.isPending,
        onClick: () => mutation.mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      {mutation.error ? (
        <Banner theme="error">
          {(mutation.error as TAPIError)?.error ||
            'Unable to revoke the secret.'}
        </Banner>
      ) : null}
      <Text>
        Requests using this secret will be rejected immediately. This action
        cannot be undone.
      </Text>
    </Modal>
  )
}

export const TriggerOverviewContainer = ({
  trigger,
}: {
  trigger: TTrigger
}) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const [revealedSecrets, setRevealedSecrets] = useState<
    Record<string, TRevealedSecret>
  >({})
  const target = {
    triggerId: trigger?.id || '',
    orgId: org?.id || '',
    triggerName: trigger?.name,
  }
  const enabled = !!target.triggerId && !!target.orgId
  const revealMutation = useMutation({
    mutationFn: (secretId: string) =>
      revealTriggerSecret({
        triggerId: target.triggerId,
        orgId: target.orgId,
        secretId,
      }),
    onSuccess: (response, secretId) => {
      if (!response?.secret) return
      setRevealedSecrets((prev) => ({
        ...prev,
        [secretId]: { secret: response.secret || '' },
      }))
    },
  })
  const eventTypes = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['event-trigger-event-types', target.orgId, target.triggerId],
    queryFn: () =>
      getTriggerEventTypes({
        triggerId: target.triggerId,
        orgId: target.orgId,
      }),
    enabled,
  })
  const ingressUrl = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['event-trigger-ingress-url', target.orgId, target.triggerId],
    queryFn: () =>
      getTriggerIngressURL({
        triggerId: target.triggerId,
        orgId: target.orgId,
      }),
    enabled,
    retry: false,
  })
  return (
    <TriggerOverview
      exampleEventType={eventTypes.data?.at(0)?.event_type}
      ingressUrl={ingressUrl.data?.ingress_url}
      ingressUrlForbidden={
        (ingressUrl.error as TAPIError | null)?.status === 403
      }
      trigger={trigger}
      revealedSecrets={revealedSecrets}
      revealError={
        revealMutation.error
          ? (revealMutation.error as TAPIError)?.error ||
            'Unable to reveal the secret.'
          : undefined
      }
      revealPendingSecretId={
        revealMutation.isPending ? revealMutation.variables : undefined
      }
      onRevealSecret={
        enabled ? (secretId) => revealMutation.mutate(secretId) : undefined
      }
      onHideSecret={(secretId) =>
        setRevealedSecrets((prev) => {
          const next = { ...prev }
          delete next[secretId]
          return next
        })
      }
      onRotateIngressURL={
        enabled
          ? () => addModal(<RotateIngressURLModal {...target} />)
          : undefined
      }
      onRotateSecret={
        enabled
          ? () => {
              setRevealedSecrets({})
              addModal(<RotateSecretModal {...target} />)
            }
          : undefined
      }
      onRevokeSecret={
        enabled
          ? (secretId) => {
              setRevealedSecrets({})
              addModal(<RevokeSecretModal {...target} secretId={secretId} />)
            }
          : undefined
      }
    />
  )
}

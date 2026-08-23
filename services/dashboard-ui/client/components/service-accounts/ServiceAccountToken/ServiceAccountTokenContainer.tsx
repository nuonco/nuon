import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createServiceAccountToken } from '@/lib'
import type { TAccount } from '@/types'
import { CreateServiceAccountTokenModal } from './ServiceAccountToken'

// Takes an account ID and a label rather than a TAccount: callers that hold a
// whole account pass its fields, and callers that only know the ID — the install
// stack tab, which resolves its service account from the stack — pass that. The
// mutation never needed more than these two values.
interface ICreateServiceAccountToken {
  accountId: string
  identity: string
  defaultDuration?: string
  onCreated?: () => void
}

export const CreateServiceAccountTokenModalContainer = ({
  accountId,
  identity,
  defaultDuration,
  onCreated,
  ...props
}: ICreateServiceAccountToken & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [createdToken, setCreatedToken] = useState<string | null>(null)

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({
      duration,
      invalidate,
    }: {
      duration: string
      invalidate: boolean
    }) =>
      createServiceAccountToken({
        body: { duration, invalidate },
        accountId,
        orgId: org.id,
      }),
    onSuccess: (data) => {
      setCreatedToken(data?.token ?? null)
      onCreated?.()
      addToast(
        <Toast heading="Token created" theme="success">
          <Text>Created a token for {identity}.</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Create failed" theme="error">
          <Text>There was an error creating a token for {identity}.</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateServiceAccountTokenModal
      accountIdentity={identity}
      defaultDuration={defaultDuration}
      isPending={isPending}
      error={error}
      createdToken={createdToken}
      onSubmit={({ duration, invalidate }) => mutate({ duration, invalidate })}
      onDone={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateServiceAccountTokenButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <CreateServiceAccountTokenModalContainer
      accountId={account.id || ''}
      identity={account.email || account.id || ''}
    />
  )

  return (
    <Button
      variant="ghost"
      className="!p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Create token
      {props?.isMenuButton ? <Icon variant="KeyIcon" /> : null}
    </Button>
  )
}

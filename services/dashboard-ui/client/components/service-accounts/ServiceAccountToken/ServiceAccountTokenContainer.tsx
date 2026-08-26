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

const CreateServiceAccountTokenModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [createdToken, setCreatedToken] = useState<string | null>(null)
  const identity = account.email || account.id || ''

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
        accountId: account.id || '',
        orgId: org.id,
      }),
    onSuccess: (data) => {
      setCreatedToken(data?.token ?? null)
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
  const modal = <CreateServiceAccountTokenModalContainer account={account} />

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

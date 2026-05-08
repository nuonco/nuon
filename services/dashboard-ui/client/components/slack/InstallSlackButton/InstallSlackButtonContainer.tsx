import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IButtonAsButton } from '@/components/common/Button'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getSlackInstallURL } from '@/lib'
import type { TAPIError } from '@/types'
import { InstallSlackButton } from './InstallSlackButton'

export const InstallSlackButtonContainer = (
  props: Omit<IButtonAsButton, 'children' | 'onClick'> = {}
) => {
  const { org } = useOrg()
  const { addToast } = useToast()

  const { mutate, isPending } = useMutation({
    mutationFn: () => getSlackInstallURL({ orgId: org.id }),
    onSuccess: ({ url }) => {
      if (url) {
        window.location.href = url
      }
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to start Slack install" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <InstallSlackButton
      isPending={isPending}
      onInstall={() => mutate()}
      {...props}
    />
  )
}

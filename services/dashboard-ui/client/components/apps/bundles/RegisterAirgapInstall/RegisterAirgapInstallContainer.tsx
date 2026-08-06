import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createAirgapInstall } from '@/lib'
import type { TAirgapBundle, TAPIError } from '@/types'
import { RegisterAirgapInstallModal } from './RegisterAirgapInstall'

export const RegisterAirgapInstallModalContainer = ({
  bundle,
  ...props
}: { bundle: TAirgapBundle } & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { error, mutate, isPending } = useMutation({
    mutationFn: (body: { name: string }) =>
      createAirgapInstall({
        appId: bundle.app_id!,
        body,
        bundleId: bundle.id!,
        orgId: org!.id,
      }),
    onSuccess: (install) => {
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({
        queryKey: ['airgap-installs', org?.id, bundle.app_id, bundle.id],
      })
      addToast(
        <Toast heading="Install registered" theme="success">
          <Text>
            <Link href={`/${org?.id}/installs/${install.id}`}>
              {install.name}
            </Link>{' '}
            now tracks this bundle delivery.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <RegisterAirgapInstallModal
      bundleId={bundle.id}
      isPending={isPending}
      error={error as TAPIError | null}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const RegisterAirgapInstallButton = ({
  bundle,
  ...props
}: { bundle: TAirgapBundle } & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => {
        addModal(<RegisterAirgapInstallModalContainer bundle={bundle} />)
      }}
      {...props}
    >
      <Icon variant="CloudSlashIcon" size={16} />
      Register install
    </Button>
  )
}

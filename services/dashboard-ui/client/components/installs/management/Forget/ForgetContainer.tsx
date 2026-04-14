import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { forgetInstall } from '@/lib'
import { ForgetModal } from './Forget'

interface IForget {}

export const ForgetModalContainer = ({ ...props }: IForget & Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      forgetInstall({
        orgId: org.id,
        installId: install.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({ queryKey: ['app-installs'] })
      addToast(
        <Toast heading="Install forgotten" theme="success">
          <Text>Install {install.name} has been forgotten.</Text>
        </Toast>
      )
      navigate(`/${org.id}/installs`)
      removeModal(props.modalId)
    },
    onError: (error) => {
      addToast(
        <Toast heading="Forget failed" theme="error">
          <Text>Unable to forget {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <ForgetModal
      installName={install.name}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const ForgetButton = ({
  isMenuButton: _isMenuButton,
  ...props
}: IForget & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ForgetModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      variant="ghost"
      className="!bg-transparent !border-0 !p-2 text-sm !leading-none h-8 w-full flex justify-between !rounded-md !text-red-800 dark:!text-red-500 hover:!bg-red-50 dark:hover:!bg-[#1D0D10] active:!bg-red-100 dark:active:!bg-[#2E1013]"
      {...props}
    >
      Forget install
      <Icon variant="Trash" />
    </Button>
  )
}

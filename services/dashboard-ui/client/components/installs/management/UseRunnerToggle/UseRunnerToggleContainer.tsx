import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { IModal } from '@/components/surfaces/Modal'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createInstallConfig, updateInstallConfig } from '@/lib'
import { UseRunnerToggleModal } from './UseRunnerToggle'

export const UseRunnerToggleModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { addToast } = useToast()

  const hasInstallConfig = Boolean(install?.install_config)
  const usesRunner = install?.install_config?.use_runner ?? true

  const {
    mutate,
    isPending: isLoading,
    data,
    error,
  } = useMutation({
    mutationFn: async () => {
      if (hasInstallConfig) {
        return updateInstallConfig({
          orgId: org.id,
          installId: install.id,
          installConfigId: install.install_config!.id!,
          body: { use_runner: !usesRunner },
        })
      } else {
        return createInstallConfig({
          orgId: org.id,
          installId: install.id,
          body: { use_runner: !usesRunner },
        })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install', org.id, install.id],
      })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      addToast(
        <Toast heading="Stack reprovision started" theme="info">
          <Text>
            Switching{' '}
            <Badge variant="code" size="md">
              {install.name}
            </Badge>{' '}
            to {usesRunner ? 'runnerless' : 'a runner'}. This may take a few
            minutes.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(
        installLink({
          installId: install.id,
          appId: install.app_id,
          suffix: '/workflows',
        })
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Switch failed" theme="error">
          <Text>
            Unable to switch{' '}
            <Badge variant="code" size="md">
              {install.name}
            </Badge>{' '}
            to {usesRunner ? 'runnerless' : 'a runner'}.
          </Text>
        </Toast>
      )
    },
  })

  return (
    <UseRunnerToggleModal
      isPending={isLoading}
      error={error}
      usesRunner={usesRunner}
      isSuccess={Boolean(data)}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

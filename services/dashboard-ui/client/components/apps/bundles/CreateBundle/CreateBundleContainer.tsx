import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createAirgapBundle, getAppConfigs } from '@/lib'
import type { TAPIError } from '@/types'
import { CreateBundleModal } from './CreateBundle'

export const CreateBundleModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: configs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: app!.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { error, mutate, isPending } = useMutation({
    mutationFn: () =>
      createAirgapBundle({
        appId: app!.id,
        body: { app_config_id: appConfigId! },
        orgId: org!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['airgap-bundles', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Creating bundle" theme="info">
          <Text>
            Publishing a bundle for{' '}
            <Badge variant="code" size="md">
              {app?.name}
            </Badge>
            . This may take a few minutes.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <CreateBundleModal
      appName={app?.name}
      appConfigId={appConfigId}
      isPending={isPending}
      error={error as TAPIError | null}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const CreateBundleButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CreateBundleModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="PackageIcon" />}
      Create bundle
      {props?.isMenuButton ? <Icon variant="PackageIcon" /> : null}
    </Button>
  )
}

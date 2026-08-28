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
import { createAppRelease, createReleasePackage, getAppConfigs } from '@/lib'
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
    mutationFn: async () => {
      const release = await createAppRelease({
        appId: app!.id,
        body: { app_config_id: appConfigId! },
        orgId: org!.id,
      })
      await createReleasePackage({
        appId: app!.id,
        body: { format: 'portable-oci', target_platform: 'linux/amd64' },
        orgId: org!.id,
        releaseId: release.id!,
      })
      return release
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-releases', org?.id, app?.id],
      })
      addToast(
        <Toast heading="Publishing release" theme="info">
          <Text>
            Publishing a release and portable bundle for{' '}
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
      Publish release
      {props?.isMenuButton ? <Icon variant="PackageIcon" /> : null}
    </Button>
  )
}

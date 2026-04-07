import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateRunner, updateMngRunner } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TRunnerSettings } from '@/types'
import {
  UpdateRunnerModal as UpdateRunnerModalComponent,
  UpdateRunnerButton as UpdateRunnerButtonComponent,
} from './UpdateRunner'

export const UpdateRunnerButton = ({
  settings,
  ...props
}: IButtonAsButton & { settings: TRunnerSettings }) => {
  const { addModal } = useSurfaces()
  const modal = <UpdateRunnerModal settings={settings} />
  return (
    <UpdateRunnerButtonComponent
      onOpenModal={() => addModal(modal)}
      {...props}
    />
  )
}

export const UpdateRunnerModal = ({
  settings,
  ...props
}: Omit<IModal, 'onSubmit'> & { settings: TRunnerSettings }) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const {
    data: isUpdated,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: async (tag: string) => {
      await updateRunner({
        runnerId: runner.id,
        orgId: org.id,
        body: {
          container_image_tag: tag,
          container_image_url: settings?.container_image_url,
          org_awsiam_role_arn: settings?.org_aws_iam_role_arn || '',
          org_k8s_service_account_name: settings?.org_k8s_service_account_name,
          runner_api_url: settings?.runner_api_url,
        },
      })

      if (runner?.runner_group?.type !== 'org') {
        await updateMngRunner({ orgId: org.id, runnerId: runner.id }).catch(() => {})
      }
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Runner update started" theme="success">
          <Text>Runner update initiated successfully.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Runner update failed" theme="error">
          <Text>Unable to update runner.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'runner_update',
        status: 'error',
        user,
        props: { orgId: org.id, runnerId: runner.id, err: error?.error },
      })
    }
    if (isUpdated as unknown) {
      trackEvent({
        event: 'runner_update',
        status: 'ok',
        user,
        props: { orgId: org.id, runnerId: runner.id },
      })
    }
  }, [isUpdated, error, org.id, runner.id, user])

  return (
    <UpdateRunnerModalComponent
      isPending={isLoading}
      error={error}
      onSubmit={(tag) => mutate(tag)}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

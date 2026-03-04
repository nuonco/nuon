import { useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateRunner, updateMngRunner } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TRunnerSettings } from '@/types'

export const UpdateRunnerButton = ({
  settings,
  ...props
}: IButtonAsButton & { settings: TRunnerSettings }) => {
  const { addModal } = useSurfaces()
  const modal = <UpdateRunnerModal settings={settings} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ArrowsCounterClockwise" />}
      Update runner version
      {props?.isMenuButton ? <Icon variant="ArrowsCounterClockwise" /> : null}
    </Button>
  )
}

export const UpdateRunnerModal = ({
  settings,
  ...props
}: IModal & { settings: TRunnerSettings }) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { runner } = useRunner()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const formRef = useRef<HTMLFormElement>(null)

  const [tag, setTag] = useState('')

  const {
    data: isUpdated,
    error,
    mutate,
    isPending: isLoading,
  } = useMutation({
    mutationFn: async () => {
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

  const handleClose = () => {
    setTag('')
    removeModal(props.modalId)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutate()
  }

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

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

  const canUpdate = tag.trim().length > 0 && !isLoading

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="ArrowsCounterClockwise" size="24" />
            Update runner version
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Updating runner
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowsCounterClockwise" />
            Update runner version
          </span>
        ),
        disabled: !canUpdate,
        onClick: handleFormSubmit,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update runner.'}
          </Banner>
        ) : null}
        <form ref={formRef} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Update to a different runner version.
            </Text>
            <div className="flex flex-col gap-2">
              <Text variant="base" weight="stronger">
                Enter the runner tag you&apos;d like to update to.
              </Text>
              <Input
                id="runner-tag"
                placeholder="runner tag"
                type="text"
                value={tag}
                onChange={(e) => setTag(e.target.value)}
                required
              />
            </div>
          </div>
        </form>
      </div>
    </Modal>
  )
}

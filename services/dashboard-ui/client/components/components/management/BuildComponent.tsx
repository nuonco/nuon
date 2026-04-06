import { useLocation, useNavigate } from 'react-router'
import { useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { buildComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'

export const BuildComponentButton = ({
  component,
  ...props
}: IButtonAsButton & {
  component: TComponent
}) => {
  const { addModal } = useSurfaces()
  const modal = <BuildComponentModal component={component} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Hammer" />}
      Build component
      {props?.isMenuButton ? <Icon variant="Hammer" /> : null}
    </Button>
  )
}

export const BuildComponentModal = ({
  component,
  ...props
}: IModal & {
  component: TComponent
}) => {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { data: build, error, mutate, isPending: isLoading } = useMutation({
    mutationFn: () => buildComponent({ componentId: component.id, orgId: org.id }),
    onSuccess: (build) => {
      addToast(
        <Toast heading={`${component.name} build started`} theme="success">
          <Text>Build for {component.name} was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      if (build?.id) {
        navigate(`${pathname}/builds/${build.id}`)
      }
    },
    onError: () => {
      addToast(
        <Toast heading="Component build failed" theme="error">
          <Text>Unable to build {component.name} component.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_build',
        status: 'error',
        user,
        props: { orgId: org.id, appId: app.id, componentId: component.id },
      })
    }
    if (build) {
      trackEvent({
        event: 'component_build',
        status: 'ok',
        user,
        props: { orgId: org.id, appId: app.id, componentId: component.id },
      })
    }
  }, [build, error, org.id, app.id, component.id, user])

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Hammer" size="24" />
          Build {component.name} component?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building component
          </span>
        ) : (
          'Build component'
        ),
        disabled: isLoading,
        onClick: () => mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to build component'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build {component.name}?
        </Text>
        <Text variant="base">
          This will start a build for the {component.name} component. The build
          process may take several minutes to complete.
        </Text>
        <Text variant="subtext" theme="neutral">
          You will be redirected to the build details page to monitor progress.
        </Text>
      </div>
    </Modal>
  )
}

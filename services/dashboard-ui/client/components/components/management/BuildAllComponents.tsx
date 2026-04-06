import { useLocation } from 'react-router'
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
import { buildComponents } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const BuildAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <BuildAllComponentsModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Build all components
      <Icon variant="Hammer" />
    </Button>
  )
}

export const BuildAllComponentsModal = ({ ...props }: IModal) => {
  const { pathname } = useLocation()
  const { user } = useAuth()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { data: buildOk, error, mutate, isPending: isLoading } = useMutation({
    mutationFn: () => buildComponents({ appId: app.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading={`${app.name} component builds started`} theme="success">
          <Text>Build all components workflow was started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Component builds failed" theme="error">
          <Text>Unable to build all components for {app.name}.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_build',
        status: 'error',
        user,
        props: { appId: app.id, orgId: org.id, err: error?.error },
      })
    }
    if (buildOk) {
      trackEvent({
        event: 'components_build',
        status: 'ok',
        user,
        props: { appId: app.id, orgId: org.id },
      })
    }
  }, [buildOk, error, app.id, org.id, user])

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
          Build all components for {app.name}?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building components
          </span>
        ) : (
          'Build components'
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
            {error?.error || 'Unable to build components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build all components?
        </Text>
        <Text variant="base">
          This will build all components in the application. This process may
          take several minutes to complete.
        </Text>
        <Text variant="subtext" theme="neutral">
          You can monitor the progress of each component build in the components
          table.
        </Text>
      </div>
    </Modal>
  )
}

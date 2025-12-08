'use client'

import { usePathname } from 'next/navigation'
import { useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { buildComponents } from '@/actions/apps/build-components'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
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
  const path = usePathname()
  const { user } = useUser()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()

  const {
    data: buildOk,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: buildComponents })

  useServerActionToast({
    data: buildOk,
    error,
    errorContent: <Text>Unable to build all components for {app.name}.</Text>,
    errorHeading: `Component builds failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Build all components workflow was started.</Text>,
    successHeading: `${app.name} component builds started`,
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_build',
        status: 'error',
        user,
        props: {
          appId: app.id,
          orgId: org.id,
          err: error?.error,
        },
      })
    }

    if (buildOk) {
      trackEvent({
        event: 'components_build',
        status: 'ok',
        user,
        props: {
          appId: app.id,
          orgId: org.id,
        },
      })
    }
  }, [buildOk, error, app.id, org.id, user])

  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
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
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            appId: app.id,
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
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

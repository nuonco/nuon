'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { buildComponent } from '@/actions/apps/build-component'
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
import type { TComponent } from '@/types'
import { trackEvent } from '@/lib/segment-analytics'

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
  const path = usePathname()
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()

  const {
    data: build,
    error,
    execute,
    isLoading,
  } = useServerAction({ action: buildComponent })

  useServerActionToast({
    data: build,
    error,
    errorContent: <Text>Unable to build {component.name} component.</Text>,
    errorHeading: `Component build failed`,
    onSuccess: () => {
      removeModal(props.modalId)
      if (build?.id) {
        const buildPath = `${path}/builds/${build.id}`
        router.push(buildPath)
      }
    },
    successContent: <Text>Build for {component.name} was started.</Text>,
    successHeading: `${component.name} build started`,
  })

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'component_build',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          appId: app.id,
          componentId: component.id,
        },
      })
    }

    if (build) {
      trackEvent({
        event: 'component_build',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          appId: app.id,
          componentId: component.id,
        },
      })
    }
  }, [build, error, org.id, app.id, component.id, user])

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
        onClick: () => {
          execute({
            componentId: component.id,
            orgId: org.id,
            path,
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
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

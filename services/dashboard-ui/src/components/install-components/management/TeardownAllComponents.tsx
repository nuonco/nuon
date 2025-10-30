'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { teardownComponents } from '@/actions/installs/teardown-components'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { trackEvent } from '@/lib/segment-analytics'

export const TeardownAllComponentsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownAllComponentsModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      Teardown components <Icon variant="CloudArrowDownIcon" />
    </Button>
  )
}

export const TeardownAllComponentsModal = ({ ...props }: IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { user } = useUser()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const [isKickedOff, setIsKickedOff] = useState(false)

  const {
    data: teardownsOk,
    error,
    execute,
    headers,
    isLoading,
  } = useServerAction({ action: teardownComponents })

  useServerActionToast({
    data: teardownsOk,
    error,
    errorContent: (
      <Text>Unabled to teardown all components to {install.name}.</Text>
    ),
    errorHeading: `Teardowns failed`,
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: <Text>Teardown all components workflow was created.</Text>,
    successHeading: `${install.name} component teardowns started`,
  })

  useEffect(() => {
    const kickoff = () => setIsKickedOff(false)

    if (isKickedOff) {
      const displayNotice = setTimeout(kickoff, 15000)

      return () => {
        clearTimeout(displayNotice)
      }
    }
  }, [isKickedOff])

  useEffect(() => {
    if (error) {
      trackEvent({
        event: 'components_teardown',
        status: 'error',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
          err: error?.error,
        },
      })
    }

    if (teardownsOk) {
      trackEvent({
        event: 'components_teardown',
        status: 'ok',
        user,
        props: {
          installId: install.id,
          orgId: org.id,
        },
      })

      if (headers?.['x-nuon-install-workflow-id']) {
        router.push(
          `/${org.id}/installs/${install.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`
        )
      } else {
        router.push(`/${org.id}/installs/${install.id}/workflows`)
      }
    }
  }, [teardownsOk, error, headers])

  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="CloudArrowDownIcon" size="24" />
          Teardown all {install.name} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting teardown
          </span>
        ) : (
          'Teardown components'
        ),
        disabled: isKickedOff,
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            installId: install.id,
            body: { plan_only: false },
          })
        },
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error?.error ? (
          <Banner theme="error">
            {error?.error || 'Unable to teardown components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to teardown all components?
        </Text>
        <Text variant="base">
          This aciton will remove all running component deployments from your
          install.
        </Text>
      </div>
    </Modal>
  )
}

'use client'

import { usePathname, useRouter } from 'next/navigation'
import { reprovisionInstall } from '@/actions/installs/reprovision-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

interface IReprovision {}

export const ReprovisionModal = ({ ...props }: IReprovision & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, error, headers, isLoading, execute } = useServerAction({
    action: reprovisionInstall,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unabled to reprovision {install.name}.</Text>,
    errorHeading: `Reprovision not started`,
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Reprovisioning {install.name} workflow was created.</Text>
    ),
    successHeading: `${install.name} reprovision was started`,
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="ArrowURightUp" size="24" />
          Reprovision install?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting reprovision
          </span>
        ) : (
          'Reprovision install'
        ),
        onClick: () => {
          // TODO(nnnnat): add plan only option
          execute({
            orgId: org.id,
            path,
            installId: install.id,
            body: { plan_only: false },
          })
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happned, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="strong">
          Are you sure you want to reprovision {}?
        </Text>
        <Text variant="base">
          Reprovisioning will recreate all resources and deploy all components
          again.
        </Text>
      </div>
    </Modal>
  )
}

export const ReprovisionButton = ({
  ...props
}: IReprovision & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ReprovisionModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Reprovision install
      <Icon variant="ArrowURightUp" />
    </Button>
  )
}

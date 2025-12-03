'use client'

import { useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { deprovisionInstall } from '@/actions/installs/deprovision-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

interface IDeprovision {}

export const DeprovisionModal = ({ ...props }: IDeprovision & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  
  const [confirmName, setConfirmName] = useState('')

  const { data, error, headers, isLoading, execute } = useServerAction({
    action: deprovisionInstall,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <Text>Unable to start deprovision workflow for {install.name}.</Text>
    ),
    errorHeading: `Deprovision failed`,
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      const base = `/${org.id}/installs/${install.id}/workflows`
      const workflowPath = workflowId ? `${base}/${workflowId}` : base
      router.push(workflowPath)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>Deprovision workflow started for {install.name}.</Text>
    ),
    successHeading: `Deprovision started`,
  })

  const isConfirmValid = confirmName === install.name
  const canDeprovision = isConfirmValid && !isLoading

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="ArrowDown" size="24" />
          Deprovision entire install
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Deprovisioning...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowDown" />
            Deprovision install
          </span>
        ),
        onClick: () => {
          execute({
            orgId: org.id,
            path,
            installId: install.id,
            body: {
              plan_only: false,
              error_behavior: 'abort',
            },
          })
        },
        disabled: !canDeprovision,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to kick off install deprovision'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to deprovision {install.name}?
            </Text>
            <Text variant="body" theme="neutral">
              Deprovisioning an install will remove it from the cloud account.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              This will create a workflow that attempts to:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Teardown each install component according to the dependency order.</li>
              <li>Teardown the install sandbox</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {install.name}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-install-name"
              placeholder="install name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Install name doesn't match" : undefined}
            />
          </div>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Important:</strong> After this workflow completes, please manually teardown the CloudFormation stack in the AWS console.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}

export const DeprovisionButton = ({
  ...props
}: IDeprovision & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeprovisionModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Deprovision install
      <Icon variant="ArrowDown" />
    </Button>
  )
}
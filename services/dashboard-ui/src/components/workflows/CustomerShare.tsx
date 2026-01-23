'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'


import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import type { TWorkflow } from '@/types'

interface ICustomerShare {
  workflow: TWorkflow
}

export const CustomerShareModal = ({
  workflow,
  ...props
}: ICustomerShare & IModal) => {
  const { install } = useInstall()
  
  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Export" size="24" />
          {`Share this ${install?.app?.name} ${workflow?.type} with your customer`}
        </Text>
      }
    
      {...props}
    >
      <div className="flex flex-col gap-1">
        
        <Text variant="base" weight="stronger">
         TKTK
        </Text>
        <Text variant="base">
          TKTKTKTK
        </Text>
      </div>
    </Modal>
  )
}


export const CustomerShareButton = ({
  workflow,
  ...props
}: ICustomerShare & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CustomerShareModal workflow={workflow} />
  const { canShowCustomerShare } = useWorkflowActions(workflow, false)

  return canShowCustomerShare ? (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Share with customer
      {props?.isMenuButton ? <Icon variant="Export" /> : null}
    </Button>
  ) : null
}

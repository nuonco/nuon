'use client'

import React, { type FC } from 'react'
import { Button, type IButton } from '@/components/Button'
import { runManualWorkflow } from './workflow-actions'

interface IActionTriggerButton extends Omit<IButton, 'className' | 'onClick'> {
  installId: string
  orgId: string
  workflowConfigId: string
}

export const ActionTriggerButton: FC<IActionTriggerButton> = ({
  installId,
  orgId,
  workflowConfigId,
  ...props
}) => {
  return (
    <Button
      className="text-sm !py-2 !h-fit"
      onClick={() => {
        runManualWorkflow({ installId, orgId, workflowConfigId })
      }}
      {...props}
    >
      Run workflow
    </Button>
  )
}

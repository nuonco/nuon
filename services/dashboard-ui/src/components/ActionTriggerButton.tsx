'use client'

import React, { type FC } from 'react'
import { Button } from '@/components/Button'
import { runManualWorkflow } from './workflow-actions'

export const ActionTriggerButton: FC<{
  installId: string
  orgId: string
  workflowConfigId: string
}> = ({ installId, orgId, workflowConfigId }) => {
  return (
    <Button
      className="text-sm !py-2 !h-fit"
      onClick={() => {
        runManualWorkflow({ installId, orgId, workflowConfigId })
      }}
    >
      Run workflow
    </Button>
  )
}

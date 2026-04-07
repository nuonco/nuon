import React, { type ReactElement, type ReactNode } from 'react'
import { Divider } from '@/components/common/Divider'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TWorkflowStep } from '@/types'
import { StepBanner } from '../StepBanner'
import { StepTitle } from '../StepTitle'
import { StepMetadata } from '../StepMetadata'

export interface IStepDetailPanel extends IPanel {
  children: ReactNode
  step: TWorkflowStep
  planOnly?: boolean
}

export const StepDetailPanel = ({
  children,
  step,
  planOnly = false,
  ...props
}: IStepDetailPanel) => {
  return (
    <Panel
      className="@container"
      heading={<StepTitle step={step} />}
      size="half"
      {...props}
    >
      <StepBanner step={step} planOnly={planOnly} />
      {React.Children.map(children, (c) =>
        React.isValidElement(c)
          ? React.cloneElement(
              c as ReactElement<{ step: TWorkflowStep; panelId: string }>,
              { step, panelId: props.panelId }
            )
          : null
      )}

      <Divider dividerWord="Metadata" />

      <StepMetadata step={step} />
    </Panel>
  )
}

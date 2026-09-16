import { useEffect, useRef, useState, type HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Button, type IButton } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import {
  resolveWizardStep,
  type IWizardDescriptor,
} from '../../../utils/wizard'
import { WizardStepper } from './WizardStepper'

export interface IWizard<TState>
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  descriptor: IWizardDescriptor<TState>
  state: TState
  exitAction?: IButton
}

export const Wizard = <TState,>({
  descriptor,
  state,
  exitAction,
  className,
  ...props
}: IWizard<TState>) => {
  const steps = descriptor.steps ?? []
  const resolved = resolveWizardStep(descriptor, state)
  const currentId = resolved.current?.id ?? steps.at(-1)?.id
  const currentIndex = steps.findIndex((step) => step.id === currentId)
  const [revisitedId, setRevisitedId] = useState<string>()
  const lastCurrentId = useRef(currentId)

  useEffect(() => {
    if (lastCurrentId.current === currentId) return
    lastCurrentId.current = currentId
    setRevisitedId(undefined)
  }, [currentId])

  const revisitedIndex = steps.findIndex((step) => step.id === revisitedId)
  const revisitedStatus =
    revisitedIndex === -1 ? undefined : resolved.statuses[revisitedIndex]
  const showingIndex =
    revisitedStatus && revisitedStatus !== 'upcoming'
      ? revisitedIndex
      : currentIndex
  const showing = steps[showingIndex]
  const readOnly = resolved.statuses[showingIndex] !== 'current'

  return (
    <div className={cn('flex w-full flex-col gap-6', className)} {...props}>
      <WizardStepper
        steps={steps.map((step, index) => ({
          id: step.id,
          label: step.label,
          status: resolved.statuses[index] ?? 'upcoming',
        }))}
        selectedId={showing?.id}
        onSelect={(id) => setRevisitedId(id === currentId ? undefined : id)}
      />
      {showing ? (
        <Card className="flex flex-col gap-4">
          {showing.render({ state, readOnly })}
        </Card>
      ) : null}
      {exitAction ? (
        <div className="flex justify-end">
          <Button
            {...exitAction}
            disabled={!resolved.finished || exitAction.disabled}
          />
        </div>
      ) : null}
    </div>
  )
}

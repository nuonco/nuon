import {
  Children,
  cloneElement,
  isValidElement,
  useState,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Icon } from '../atoms/Icon'
import { DisclosureGroup, ExpandAllButton } from '../molecules/DisclosureGroup'
import type { TDiffView } from '../molecules/Diff'
import { DiffSection, type IDiffSection } from './DiffSection'

export interface IDiffSections
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  children: ReactNode
  defaultOpen?: boolean
  defaultView?: TDiffView
}

interface IDiffControls {
  view: TDiffView
  setView: (view: TDiffView) => void
}

const DiffControls = ({ view, setView }: IDiffControls) => {
  const split = view === 'split'
  return (
    <div
      aria-label="Diff controls"
      className="flex flex-wrap items-center justify-end gap-1 pb-2"
    >
      <ExpandAllButton />
      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-pressed={split}
        aria-label={split ? 'Unified view' : 'Split view'}
        tooltip={split ? 'Unified view' : 'Split view'}
        onClick={() => setView(split ? 'unified' : 'split')}
      >
        <Icon
          variant={
            split ? 'SquareSplitVerticalIcon' : 'SquareSplitHorizontalIcon'
          }
          size={14}
        />
      </Button>
    </div>
  )
}

export const DiffSections = ({
  children,
  defaultOpen = false,
  defaultView = 'unified',
  className,
  ...props
}: IDiffSections) => {
  const [view, setView] = useState<TDiffView>(defaultView)
  const sections = Children.map(children, (child) =>
    isValidElement<IDiffSection>(child) && child.type === DiffSection
      ? cloneElement(child, { view })
      : child
  )

  return (
    <DisclosureGroup
      defaultOpen={defaultOpen}
      className={cn('gap-1', className)}
      {...props}
    >
      <DiffControls view={view} setView={setView} />
      {sections}
    </DisclosureGroup>
  )
}

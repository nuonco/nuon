import {
  Children,
  cloneElement,
  isValidElement,
  useState,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useDisclosureGroup } from '../../hooks/use-disclosure'
import { Button } from '../atoms/Button'
import { Icon } from '../atoms/Icon'
import { DisclosureGroup, ExpandAllButton } from '../molecules/DisclosureGroup'
import type { TDiffView } from '../molecules/Diff'
import { DiffSection, type IDiffSection } from './DiffSection'

export interface IDiffSections
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  children: ReactNode
  toolbar?: ReactNode
  defaultOpen?: boolean
  defaultView?: TDiffView
}

interface IDiffControls {
  view: TDiffView
  setView: (view: TDiffView) => void
}

const DiffControls = ({ view, setView }: IDiffControls) => {
  const group = useDisclosureGroup()
  const split = view === 'split'

  if (!group?.count) return null

  return (
    <div
      aria-label="Diff controls"
      className="ml-auto flex items-center gap-0.5"
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
  toolbar,
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
      <div className="flex flex-wrap items-center gap-4 pb-2">
        {toolbar}
        <DiffControls view={view} setView={setView} />
      </div>
      {sections}
    </DisclosureGroup>
  )
}

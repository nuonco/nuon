import {
  Children,
  cloneElement,
  isValidElement,
  useEffect,
  useState,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useDashboardPreferences } from '@/hooks/use-dashboard-preferences'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useDisclosureGroup } from './use-disclosure'
import { DisclosureGroup, ExpandAllButton } from './DisclosureGroup'
import type { TDiffView } from './Diff'
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
  divider?: boolean
}

const DiffControls = ({ view, setView, divider = false }: IDiffControls) => {
  const group = useDisclosureGroup()
  const split = view === 'split'

  if (!group?.count) return null

  return (
    <div
      aria-label="Diff controls"
      className="ml-auto flex items-center gap-0.5"
    >
      {divider ? (
        <span aria-hidden className="mx-1.5 h-4 border-l" />
      ) : null}
      <ExpandAllButton />
      <Button
        size="sm"
        variant="icon"
        aria-pressed={split}
        aria-label={split ? 'Unified view' : 'Split view'}
        tooltipProps={{ tipContent: split ? 'Unified view' : 'Split view' }}
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
  defaultOpen,
  defaultView,
  className,
  ...props
}: IDiffSections) => {
  const { diffView, planSections } = useDashboardPreferences()
  const [localView, setLocalView] = useState<TDiffView>(
    defaultView ?? diffView
  )

  useEffect(() => {
    if (defaultView === undefined) setLocalView(diffView)
  }, [defaultView, diffView])

  const sections = Children.map(children, (child) =>
    isValidElement<IDiffSection>(child) && child.type === DiffSection
      ? cloneElement(child, { view: localView })
      : child
  )

  return (
    <DisclosureGroup
      defaultOpen={defaultOpen ?? planSections === 'expanded'}
      className={cn('gap-1', className)}
      {...props}
    >
      <div className="flex flex-wrap items-center gap-2 pb-2">
        {toolbar}
        <DiffControls
          view={localView}
          setView={setLocalView}
          divider={!!toolbar}
        />
      </div>
      {sections}
    </DisclosureGroup>
  )
}

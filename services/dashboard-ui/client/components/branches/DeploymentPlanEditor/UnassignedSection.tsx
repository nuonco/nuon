import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { Text } from '@/components/common/Text'
import type { TInstall } from '@/types'
import { cn } from '@/utils/classnames'
import { SortableInstallRow } from './SortableInstallRow'

interface IUnassignedSection {
  installs: TInstall[]
  containerId: string
  disabled?: boolean
}

export const UnassignedSection = ({
  installs,
  containerId,
  disabled,
}: IUnassignedSection) => {
  const { setNodeRef, isOver } = useDroppable({
    id: containerId,
    data: { containerId, type: 'container' },
  })

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <Text variant="subtext" weight="strong" theme="neutral">
          Unassigned
        </Text>
        <Text variant="subtext" theme="neutral">
          {installs.length === 0
            ? '— all installs are assigned'
            : `— ${installs.length} install${installs.length === 1 ? '' : 's'} won't deploy`}
        </Text>
      </div>

      <div
        ref={setNodeRef}
        className={cn(
          'rounded-md transition-colors',
          isOver && 'bg-primary-50/40 dark:bg-primary-900/10'
        )}
      >
        <SortableContext
          items={installs.map((i) => i.id)}
          strategy={verticalListSortingStrategy}
        >
          {installs.length > 0 ? (
            <div className="flex flex-col gap-1">
              {installs.map((install) => (
                <SortableInstallRow
                  key={install.id}
                  installId={install.id}
                  installName={install.name || install.id}
                  containerId={containerId}
                  disabled={disabled}
                />
              ))}
            </div>
          ) : null}
        </SortableContext>
      </div>
    </div>
  )
}

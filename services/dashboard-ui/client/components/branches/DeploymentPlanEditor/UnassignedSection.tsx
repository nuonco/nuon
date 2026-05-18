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
    <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg bg-white dark:bg-dark-grey-800">
      <div className="grid grid-cols-1 md:grid-cols-[minmax(220px,280px)_1fr] divide-y md:divide-y-0 md:divide-x divide-cool-grey-200 dark:divide-dark-grey-700">
        <div className="flex flex-col gap-1 p-4">
          <Text variant="base" weight="strong">
            Unassigned
          </Text>
          <Text variant="subtext" theme="neutral">
            {installs.length === 0
              ? 'All installs are assigned to a group.'
              : "These installs won't deploy. Drag them into a group."}
          </Text>
        </div>

        <div
          ref={setNodeRef}
          className={cn(
            'flex flex-col gap-2 p-4 transition-colors',
            isOver && 'bg-primary-50/40 dark:bg-primary-900/10'
          )}
        >
          <SortableContext
            items={installs.map((i) => i.id)}
            strategy={verticalListSortingStrategy}
          >
            {installs.length > 0 ? (
              <div className="flex flex-col gap-1.5">
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
            ) : (
              <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
                <Text variant="subtext" theme="neutral">
                  No unassigned installs
                </Text>
              </div>
            )}
          </SortableContext>
        </div>
      </div>
    </div>
  )
}

import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface ISortableInstallRow {
  installId: string
  installName: string
  containerId: string
  disabled?: boolean
  showRemove?: boolean
  onRemove?: () => void
}

export const SortableInstallRow = ({
  installId,
  installName,
  containerId,
  disabled,
  showRemove,
  onRemove,
}: ISortableInstallRow) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: installId,
    data: { containerId, type: 'install' },
    disabled,
  })

  const style = {
    transform: CSS.Translate.toString(transform),
    transition,
    opacity: isDragging ? 0 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'flex items-center justify-between gap-2 px-2 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900',
        !isDragging && 'cursor-default'
      )}
    >
      <div className="flex items-center gap-1.5 min-w-0">
        <button
          type="button"
          {...attributes}
          {...listeners}
          disabled={disabled}
          aria-label={`Drag ${installName}`}
          title="Drag to move"
          className={cn(
            'flex items-center justify-center p-1 rounded text-cool-grey-500 hover:text-cool-grey-800 dark:hover:text-cool-grey-300 shrink-0',
            disabled ? 'cursor-not-allowed' : 'cursor-grab active:cursor-grabbing'
          )}
        >
          <Icon variant="DotsSixVerticalIcon" size={16} />
        </button>
        <Text variant="body" className="truncate">
          {installName}
        </Text>
      </div>
      {showRemove && onRemove && (
        <Button
          variant="ghost"
          size="xs"
          onClick={onRemove}
          disabled={disabled}
          title={`Remove ${installName}`}
          aria-label={`Remove ${installName}`}
          className="!p-1 shrink-0"
        >
          <Icon variant="XIcon" size={14} />
        </Button>
      )}
    </div>
  )
}

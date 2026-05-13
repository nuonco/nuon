import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TInstall } from '@/types'

interface IInstallCard {
  install?: TInstall
  isDragging?: boolean
  isDisabled?: boolean
}

export const InstallCard = ({ install, isDragging, isDisabled }: IInstallCard) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging,
  } = useSortable({
    id: install?.id ?? '',
    disabled: isDisabled || !install,
  })

  if (!install) return null

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        'flex items-center gap-3 px-4 py-2.5 min-w-[280px] rounded-md border-2 transition-all cursor-grab active:cursor-grabbing',
        {
          'border-cool-grey-300 dark:border-dark-grey-600 bg-white dark:bg-dark-grey-800 hover:border-cool-grey-400 dark:hover:border-dark-grey-500 hover:shadow-md':
            !isDragging && !isSortableDragging && !isDisabled,
          'border-primary-500 dark:border-primary-400 bg-primary-900/10 dark:bg-primary-900/20 shadow-lg': isDragging || isSortableDragging,
          'border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-100 dark:bg-dark-grey-900 opacity-50 cursor-not-allowed': isDisabled,
        }
      )}
    >
      <Icon variant="CloudIcon" size={16} />
      <Text variant="base" className="truncate flex-1">
        {install.name}
      </Text>
    </div>
  )
}

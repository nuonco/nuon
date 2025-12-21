'use client'

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
  if (!install) return null

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging,
  } = useSortable({
    id: install.id,
    disabled: isDisabled,
  })

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
          'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 hover:border-gray-400 dark:hover:border-gray-500 hover:shadow-md':
            !isDragging && !isSortableDragging && !isDisabled,
          'border-blue-400 dark:border-blue-500 bg-blue-50 dark:bg-blue-900/30 shadow-lg': isDragging || isSortableDragging,
          'border-gray-200 dark:border-gray-700 bg-gray-100 dark:bg-gray-900 opacity-50 cursor-not-allowed': isDisabled,
        }
      )}
    >
      <Icon variant="Cloud" size={16} />
      <Text variant="base" className="truncate flex-1">
        {install.name}
      </Text>
    </div>
  )
}
import { cn } from '@/utils/classnames'
import { getInitials } from '@/utils/string-utils'

type TOrgAvatarSize = 'sm' | 'md' | 'lg' | 'xl'

const SIZE_CLASSES: Record<TOrgAvatarSize, string> = {
  sm: 'h-7 w-7 text-xs',
  md: 'h-9 w-9 text-sm',
  lg: 'h-11 w-11 text-base',
  xl: 'h-14 w-14 text-lg',
}

export interface IOrgAvatar extends React.HTMLAttributes<HTMLSpanElement> {
  name?: string
  size?: TOrgAvatarSize
}

export const OrgAvatar = ({
  name = '',
  size = 'md',
  className,
  ...props
}: IOrgAvatar) => {
  return (
    <span
      className={cn(
        'flex-none flex items-center justify-center rounded-md font-semibold text-white overflow-hidden select-none',
        'bg-cool-grey-600 dark:bg-cool-grey-500',
        SIZE_CLASSES[size],
        className
      )}
      {...props}
    >
      {getInitials(name)}
    </span>
  )
}

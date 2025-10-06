import Image from 'next/image'
import { cn } from '@/utils/classnames'
import { getInitials } from '@/utils/string-utils'

type TAvatarSizeKey = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
type TAvatarSize = { tw: string; px: number }

const AVATAR_SIZES: Record<TAvatarSizeKey, TAvatarSize> = {
  xs: { tw: 'h-6 w-6 text-xs', px: 24 },
  sm: { tw: 'h-7 w-7 text-sm', px: 28 },
  md: { tw: 'h-8 w-8 text-base', px: 32 },
  lg: { tw: 'h-9 w-9 text-lg', px: 36 },
  xl: { tw: 'h-10 w-10 text-xl', px: 40 },
}

interface IAvatarProps
  extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'children'> {
  isLoading?: boolean
  size?: TAvatarSizeKey
}

type TAvatar =
  | {
      name?: string
      src?: never
      alt?: never
    }
  | {
      alt?: string
      name?: never
      src?: string
    }

export type IAvatar = IAvatarProps & TAvatar

export const Avatar = ({
  alt = '',
  className,
  isLoading = false,
  name,
  src,
  size = 'md',
  ...props
}: IAvatar) => {
  const sizeConf = AVATAR_SIZES[size]

  return (
    <span
      className={cn(
        'flex-none flex items-center justify-center rounded-md font-sans overflow-hidden transition-all',

        isLoading
          ? 'bg-cool-grey-400 dark:bg-dark-grey-400 animate-pulse text-cool-grey-600 dark:text-white/50'
          : 'bg-cool-grey-200 dark:bg-dark-grey-300 text-cool-grey-600 dark:text-white/50',

        sizeConf.tw,
        className
      )}
      {...props}
    >
      {isLoading ? null : src ? (
        <Image
          height={sizeConf.px}
          width={sizeConf.px}
          src={src}
          alt={alt || ''}
          className="h-full w-full object-cover"
        />
      ) : (
        getInitials(name)
      )}
    </span>
  )
}

import { useEffect, useState, type HTMLAttributes } from 'react'
import { getInitials } from '@/utils/string-utils'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'

export type TAvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
export type TAvatarShape = 'circle' | 'rounded'

export interface IAvatar
  extends Omit<HTMLAttributes<HTMLSpanElement>, 'children'> {
  alt?: string
  name?: string
  src?: string
  size?: TAvatarSize
  shape?: TAvatarShape
  loading?: boolean
}

const SIZE_CLASSES: Record<TAvatarSize, string> = {
  xs: 'size-5 text-label',
  sm: 'size-7 text-caption',
  md: 'size-8 text-body',
  lg: 'size-10 text-heading',
  xl: 'size-12 text-title',
}

const ICON_SIZES: Record<TAvatarSize, number> = {
  xs: 12,
  sm: 16,
  md: 18,
  lg: 22,
  xl: 26,
}

export const Avatar = ({
  alt,
  name,
  src,
  size = 'md',
  shape = 'circle',
  loading = false,
  className,
  ...props
}: IAvatar) => {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setLoaded(false)
    setFailed(false)
  }, [src])

  const showImage = Boolean(src && !failed)
  const initials = getInitials(name)

  return (
    <span
      role={alt ? 'img' : undefined}
      aria-label={alt}
      aria-hidden={alt ? undefined : true}
      className={cn(
        'relative inline-flex shrink-0 items-center justify-center overflow-hidden bg-surface-accent font-sans font-medium text-accent',
        shape === 'circle' ? 'rounded-full' : 'rounded-lg',
        SIZE_CLASSES[size],
        loading && 'animate-pulse bg-surface-03 text-transparent',
        className
      )}
      {...props}
    >
      {loading ? null : (
        <>
          {showImage ? (
            <img
              src={src}
              alt=""
              referrerPolicy="no-referrer"
              className={cn(
                'absolute inset-0 size-full object-cover',
                !loaded && 'invisible'
              )}
              onLoad={() => setLoaded(true)}
              onError={() => setFailed(true)}
            />
          ) : null}
          {!showImage || !loaded ? (
            initials ? (
              initials
            ) : (
              <Icon variant="UserIcon" size={ICON_SIZES[size]} />
            )
          ) : null}
        </>
      )}
    </span>
  )
}

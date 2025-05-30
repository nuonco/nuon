import classNames from 'classnames'
import Image from 'next/image'
import React, { type FC } from 'react'
import { initialsFromString } from '@/utils'
import './Avatar.css'

type TAvatarSizeKey = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
type TAvatarSize = { tw: 6 | 7 | 8 | 9 | 10; px: 24 | 28 | 32 | 36 | 40 }

const AVATAR_SIZES: Record<TAvatarSizeKey, TAvatarSize> = {
  xs: { tw: 6, px: 24 },
  sm: { tw: 7, px: 28 },
  md: { tw: 8, px: 32 },
  lg: { tw: 9, px: 36 },
  xl: { tw: 10, px: 40 },
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

export const Avatar: FC<IAvatar> = ({
  alt = '',
  className,
  isLoading = false,
  name,
  src,
  size = 'md',
}) => {
  return (
    <span
      className={classNames('avatar', {
        loading: isLoading,
        [`size-${AVATAR_SIZES[size].tw}`]: true,
        [`${className}`]: Boolean(className),
      })}
    >
      {isLoading ? null : src ? (
        <Image
          height={AVATAR_SIZES[size].px}
          width={AVATAR_SIZES[size].px}
          src={src}
          alt={alt || ''}
        />
      ) : (
        initialsFromString(name)
      )}
    </span>
  )
}

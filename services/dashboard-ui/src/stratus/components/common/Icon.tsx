'use client'

import * as PhosphorIcons from '@phosphor-icons/react'
import { ComponentProps, ElementType } from 'react'
import { FaGithub } from 'react-icons/fa'

const customIcons = {
  GitHub: FaGithub,
} as const

type CustomIconVariant = keyof typeof customIcons

type PhosphorIconVariant = keyof Omit<
  typeof PhosphorIcons,
  'Icon' | 'IconContext' | 'IconBase' | 'createComponent' | 'IconProps'
>

export type TIconVariant = PhosphorIconVariant | CustomIconVariant

type PhosphorIconProps = ComponentProps<typeof PhosphorIcons.HouseIcon>

interface IconProps extends Omit<PhosphorIconProps, 'ref'> {
  variant: TIconVariant
}

export const Icon = ({
  variant,
  size = 16,
  weight = 'regular',
  ...props
}: IconProps) => {
  if (variant in customIcons) {
    const CustomIcon = customIcons[variant as CustomIconVariant]
    return <CustomIcon size={size} {...props} />
  }

  const IconComponent = PhosphorIcons[
    variant as PhosphorIconVariant
  ] as ElementType
  return <IconComponent size={size} weight={weight} {...props} />
}

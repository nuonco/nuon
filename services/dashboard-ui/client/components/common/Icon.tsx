'use client'

import { ComponentProps, ElementType } from 'react'
import { FaAws, FaGithub } from 'react-icons/fa'
import {
  SiAwslambda,
  SiDocker,
  SiGooglecloud,
  SiHelm,
  SiKubernetes,
  SiOpencontainersinitiative,
  SiTerraform,
} from 'react-icons/si'
import { VscAzure } from 'react-icons/vsc'
import * as PhosphorIcons from '@phosphor-icons/react'
import { Loading } from './Loading'

const customIcons = {
  AWS: FaAws,
  AWSLambda: SiAwslambda,
  Azure: VscAzure,
  Docker: SiDocker,
  GCP: SiGooglecloud,
  GitHub: FaGithub,
  Helm: SiHelm,
  Kubernetes: SiKubernetes,
  Loading: Loading,
  OCI: SiOpencontainersinitiative,
  Terraform: SiTerraform,
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
    return <CustomIcon size={size} className={props.className} />
  }

  const IconComponent = PhosphorIcons[variant as PhosphorIconVariant] as
    | ElementType
    | undefined

  if (!IconComponent) {
    // Optional: Show a warning or fallback icon
    return null
  }

  return <IconComponent size={size} weight={weight} {...props} />
}

import type { TComponentType } from '@/types'
import { cn } from '@/utils/classnames'
import { Brand, type TBrandVariant } from '../atoms/Brand'
import { Icon } from '../atoms/Icon'
import { Text, type IText } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export type TComponentTypeValue = TComponentType | 'pulumi_module' | 'unknown'
export type TComponentTypeDisplay = 'abbr' | 'name' | 'icon'

export interface IComponentType extends Omit<IText, 'children'> {
  type: TComponentTypeValue
  display?: TComponentTypeDisplay
  tone?: 'color' | 'mono'
  iconSize?: number | string
}

interface IComponentTypeConfig {
  abbr: string
  brand?: TBrandVariant
  name: string
}

const COMPONENT_TYPES: Record<TComponentTypeValue, IComponentTypeConfig> = {
  docker_build: {
    abbr: 'Docker',
    brand: 'Docker',
    name: 'Docker',
  },
  external_image: {
    abbr: 'OCI',
    brand: 'OCI',
    name: 'External image',
  },
  helm_chart: {
    abbr: 'Helm',
    brand: 'Helm',
    name: 'Helm',
  },
  terraform_module: {
    abbr: 'TF',
    brand: 'Terraform',
    name: 'Terraform',
  },
  job: {
    abbr: 'Job',
    brand: 'Lambda',
    name: 'Lambda',
  },
  pulumi_module: {
    abbr: 'Pulumi',
    brand: 'Pulumi',
    name: 'Pulumi',
  },
  kubernetes_manifest: {
    abbr: 'K8s',
    brand: 'Kubernetes',
    name: 'Kubernetes manifest',
  },
  pulumi: {
    abbr: 'Pulumi',
    brand: 'Pulumi',
    name: 'Pulumi',
  },
  unknown: {
    abbr: 'Unknown',
    name: 'Unknown',
  },
}

const configFor = (type: TComponentTypeValue) =>
  COMPONENT_TYPES[type] ?? COMPONENT_TYPES.unknown

export const componentTypeName = (type: TComponentTypeValue) =>
  configFor(type).name

const ComponentMark = ({
  brand,
  size,
  tone,
}: {
  brand?: TBrandVariant
  size: number | string
  tone: 'color' | 'mono'
}) =>
  brand ? (
    <Brand variant={brand} size={size} tone={tone} />
  ) : (
    <Icon variant="QuestionIcon" size={size} />
  )

export const ComponentType = ({
  type,
  display = 'name',
  tone = 'color',
  iconSize = '1em',
  className,
  ...props
}: IComponentType) => {
  const config = configFor(type)

  if (display === 'icon') {
    return (
      <Tooltip content={config.name} tabIndex={0} aria-label={config.name}>
        <ComponentMark brand={config.brand} size={iconSize} tone={tone} />
      </Tooltip>
    )
  }

  return (
    <Text
      className={cn(
        'inline-flex items-center gap-2 whitespace-nowrap',
        className
      )}
      {...props}
    >
      <ComponentMark brand={config.brand} size={iconSize} tone={tone} />
      <span>{display === 'name' ? config.name : config.abbr}</span>
    </Text>
  )
}

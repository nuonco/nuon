import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text, type IText } from '@/components/common/Text'
import type { TComponentType } from '@/types'

interface IComponentType extends Omit<IText, 'children'> {
  colorVariant?: 'mono' | 'color'
  displayVariant?: 'abbr' | 'name' | 'icon-only'
  iconSize?: string
  type: TComponentType
}

interface IComponentTypeConfig {
  abbr: string
  brandColorClass: string
  icon: TIconVariant
  name: string
}

const COMPONENT_TYPE_CONFIG: Record<
  TComponentType | 'unknown',
  IComponentTypeConfig
> = {
  docker_build: {
    abbr: 'Docker',
    brandColorClass: 'text-[#2496ED] dark:text-[#56B4F9]',
    icon: 'Docker',
    name: 'Docker',
  },
  external_image: {
    abbr: 'OCI',
    brandColorClass: 'text-[#262261] dark:text-[#8B87D1]',
    icon: 'OCI',
    name: 'External image',
  },
  helm_chart: {
    abbr: 'Helm',
    brandColorClass: 'text-[#0F1689] dark:text-[#6A70D6]',
    icon: 'Helm',
    name: 'Helm',
  },
  terraform_module: {
    abbr: 'TF',
    brandColorClass: 'text-[#7B42BC] dark:text-[#A878E0]',
    icon: 'Terraform',
    name: 'Terraform',
  },
  job: {
    abbr: 'Job',
    brandColorClass: 'text-[#FF9900] dark:text-[#FFB340]',
    icon: 'AWSLambda',
    name: 'Lambda',
  },
  kubernetes_manifest: {
    abbr: 'K8s',
    brandColorClass: 'text-[#326CE5] dark:text-[#5A8DEF]',
    icon: 'Kubernetes',
    name: 'Kubernetes manifest',
  },
  unknown: {
    abbr: 'Unknown',
    brandColorClass: '',
    icon: 'Question',
    name: 'Unknown',
  },
} as const

export const ComponentType = ({
  colorVariant = 'mono',
  type: configType,
  displayVariant = 'name',
  iconSize,
  ...props
}: IComponentType) => {
  const config =
    COMPONENT_TYPE_CONFIG[configType] || COMPONENT_TYPE_CONFIG.unknown
  const isIconOnly = displayVariant === 'icon-only'

  return (
    <Text
      className="!flex items-center gap-1 text-nowrap"
      {...props}
      title={isIconOnly ? config.name : undefined}
    >
      <Icon
        variant={config.icon}
        aria-hidden="true"
        size={iconSize}
        className={
          colorVariant === 'color' ? config.brandColorClass : undefined
        }
      />
      {!isIconOnly && (
        <span>{displayVariant === 'name' ? config.name : config.abbr}</span>
      )}
    </Text>
  )
}

import { Icon } from '@/components/common/Icon'
import { Text, type IText } from '@/components/common/Text'
import type { TComponentType } from '@/types'

interface IComponentType extends Omit<IText, 'children'> {
  displayVariant?: 'abbr' | 'name' | 'icon-only'
  type: TComponentType
}

interface IComponentTypeConfig {
  abbr: string
  icon: React.ReactElement
  name: string
}

const COMPONENT_TYPE_CONFIG: Record<
  TComponentType | 'unknown',
  IComponentTypeConfig
> = {
  docker_build: {
    abbr: 'Docker',
    icon: <Icon variant="Docker" aria-hidden="true" />,
    name: 'Docker',
  },
  external_image: {
    abbr: 'OCI',
    icon: <Icon variant="OCI" aria-hidden="true" />,
    name: 'External image',
  },
  helm_chart: {
    abbr: 'Helm',
    icon: <Icon variant="Helm" aria-hidden="true" />,
    name: 'Helm',
  },
  terraform_module: {
    abbr: 'TF',
    icon: <Icon variant="Terraform" aria-hidden="true" />,
    name: 'Terraform',
  },
  job: {
    abbr: 'Job',
    icon: <Icon variant="AWSLambda" aria-hidden="true" />,
    name: 'Lambda',
  },
  kubernetes_manifest: {
    abbr: 'K8s',
    icon: <Icon variant="Kubernetes" aria-hidden="true" />,
    name: 'Kubernetes manifest',
  },
  unknown: {
    abbr: 'Unknown',
    icon: <Icon variant="Question" aria-hidden="true" />,
    name: 'Unknown',
  },
} as const

export const ComponentType = ({
  type: configType,
  displayVariant = 'name',
  ...props
}: IComponentType) => {
  const config =
    COMPONENT_TYPE_CONFIG[configType] || COMPONENT_TYPE_CONFIG.unknown
  const isIconOnly = displayVariant === 'icon-only'

  return (
    <Text
      className="flex items-center gap-1 text-nowrap"
      {...props}
      title={isIconOnly ? config.name : undefined}
    >
      {config.icon}
      {!isIconOnly && (
        <span>{displayVariant === 'name' ? config.name : config.abbr}</span>
      )}
    </Text>
  )
}

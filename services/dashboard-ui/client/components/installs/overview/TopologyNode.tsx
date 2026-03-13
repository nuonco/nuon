import { useNavigate } from 'react-router'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { getStatusTheme } from '@/utils/status-utils'

type TComponentType =
  | 'terraform_module'
  | 'helm_chart'
  | 'docker_build'
  | 'external_image'
  | 'job'
  | 'kubernetes_manifest'
  | 'unknown'

const COMPONENT_TYPE_CONFIG: Record<
  TComponentType,
  { icon: TIconVariant; label: string }
> = {
  helm_chart: { icon: 'Helm', label: 'Helm' },
  terraform_module: { icon: 'Terraform', label: 'Terraform' },
  docker_build: { icon: 'Docker', label: 'Docker' },
  external_image: { icon: 'Package', label: 'Container' },
  kubernetes_manifest: { icon: 'Kubernetes', label: 'K8s Manifest' },
  job: { icon: 'Terminal', label: 'Job' },
  unknown: { icon: 'Cube', label: 'Unknown' },
}

const VARIANT_ICON: Record<string, TIconVariant> = {
  stack: 'StackSimple',
  sandbox: 'Cube',
}

const VARIANT_WIDTH: Record<string, string> = {
  stack: 'w-[240px]',
  sandbox: 'w-[240px]',
  component: 'w-[210px]',
}

interface ITopologyNodeProps {
  variant: 'stack' | 'sandbox' | 'component'
  name: string
  subtitle?: string
  status: string
  componentType?: string
  hasDrift?: boolean
  href: string
}

export function TopologyNode({
  variant,
  name,
  subtitle,
  status,
  componentType,
  hasDrift,
  href,
}: ITopologyNodeProps) {
  const navigate = useNavigate()
  const theme = getStatusTheme(status)

  const typeConfig =
    variant === 'component' && componentType
      ? COMPONENT_TYPE_CONFIG[componentType as TComponentType] ??
        COMPONENT_TYPE_CONFIG.unknown
      : null

  const variantIcon = VARIANT_ICON[variant]

  const borderThemeClass =
    theme === 'error'
      ? '!border-red-400 dark:!border-red-500/50'
      : theme === 'warn'
        ? '!border-orange-400 dark:!border-orange-500/50'
        : theme === 'success'
          ? '!border-green-400 dark:!border-green-500/40'
          : ''

  return (
    <button
      type="button"
      onClick={() => navigate(href)}
      className={cn(
        'relative flex flex-col gap-2 p-4 border rounded-lg shadow-sm',
        'bg-background',
        'hover:shadow-md hover:!border-primary-400 dark:hover:!border-primary-600',
        'transition-all duration-150 cursor-pointer text-left',
        VARIANT_WIDTH[variant],
        borderThemeClass
      )}
    >
      {hasDrift && (
        <div className="absolute -top-1.5 -right-1.5">
          <Icon
            variant="WarningCircle"
            size={18}
            weight="fill"
            className="text-orange-500"
          />
        </div>
      )}

      <div className="flex items-center gap-1.5">
        {variantIcon && (
          <Icon
            variant={variantIcon}
            size={16}
            weight="bold"
            className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0"
          />
        )}
        {typeConfig && (
          <Icon
            variant={typeConfig.icon}
            size={16}
            weight="bold"
            className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0"
          />
        )}
        <Text variant="body" weight="strong" className="truncate">
          {name}
        </Text>
      </div>

      {subtitle && (
        <Text variant="subtext" theme="neutral" className="truncate">
          {subtitle}
        </Text>
      )}

      <Status status={status} variant="badge" />
      {['active', 'in-progress', 'provisioning', 'building', 'applying', 'planning', 'checking-plan', 'generating', 'retrying', 'deleting'].includes(status) && (
        <svg
          className="animate-spin h-5 w-5 text-blue-600 dark:text-blue-400 absolute bottom-2 right-2 shrink-0"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle className="opacity-25" cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="2" />
          <path className="opacity-75" stroke="currentColor" strokeWidth="2" strokeLinecap="round" d="M4 12a8 8 0 018-8" />
        </svg>
      )}
    </button>
  )
}

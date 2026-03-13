import { useNavigate } from 'react-router'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { getStatusTheme } from '@/utils/status-utils'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'

type TComponentType =
  | 'terraform_module'
  | 'helm_chart'
  | 'docker_build'
  | 'external_image'
  | 'job'
  | 'kubernetes_manifest'
  | 'unknown'

const COMPONENT_TYPE_CONFIG: Record<TComponentType, { icon: TIconVariant; label: string }> = {
  helm_chart:          { icon: 'Helm',       label: 'Helm Chart' },
  terraform_module:    { icon: 'Terraform',  label: 'Terraform' },
  docker_build:        { icon: 'Docker',     label: 'Docker Build' },
  external_image:      { icon: 'Package',    label: 'Container Image' },
  kubernetes_manifest: { icon: 'Kubernetes', label: 'K8s Manifest' },
  job:                 { icon: 'Terminal',   label: 'Job' },
  unknown:             { icon: 'Cube',       label: 'Component' },
}

const VARIANT_ICON: Record<string, TIconVariant> = {
  stack:   'StackSimple',
  sandbox: 'BoundingBox',
}


// Status dot — uses design system scale
const THEME_DOT: Record<string, string> = {
  success: 'bg-green-600  dark:bg-green-400',
  error:   'bg-red-500    dark:bg-red-400',
  warn:    'bg-orange-500 dark:bg-orange-400',
  info:    'bg-blue-500   dark:bg-blue-400',
  neutral: 'bg-cool-grey-400 dark:bg-dark-grey-500',
}

// Status text — uses design system scale
const THEME_TEXT: Record<string, string> = {
  success: 'text-green-700  dark:text-green-400',
  error:   'text-red-600    dark:text-red-400',
  warn:    'text-orange-600 dark:text-orange-400',
  info:    'text-blue-600   dark:text-blue-400',
  neutral: 'text-cool-grey-600 dark:text-cool-grey-500',
}

const RUNNING_STATUSES = new Set([
  'in-progress', 'provisioning', 'building', 'applying', 'planning',
  'checking-plan', 'generating', 'retrying', 'deleting', 'executing',
  'waiting', 'started', 'syncing', 'deploying',
])

const APPROVAL_STATUSES = new Set([
  'pending-approval', 'approval-awaiting', 'awaiting-approval',
  'awaiting-install-stack-run', 'awaiting-user-run',
])

// Human-readable overrides for technical status strings
const STATUS_LABEL_OVERRIDE: Record<string, string> = {
  'awaiting-install-stack-run': 'Waiting for customer',
  'awaiting-user-run':          'Waiting for customer',
  'not-attempted':              'Not started',
  'pending-approval':           'Awaiting approval',
  'approval-awaiting':          'Awaiting approval',
  'awaiting-approval':          'Awaiting approval',
  'approval-denied':            'Approval denied',
  'checking-plan':              'Checking plan',
  'auto-skipped':               'Skipped',
  'user-skipped':               'Skipped',
  'access-error':               'Access error',
  'not-connected':              'Not connected',
  'timed-out':                  'Timed out',
  'policy_failed':              'Policy failed',
  'no-drift':                   'No drift',
}

interface ITopologyNodeProps {
  variant: 'stack' | 'sandbox' | 'component'
  name: string
  status: string
  statusDescription?: string
  componentType?: string
  hasDrift?: boolean
  href: string
}

export function TopologyNode({
  variant,
  name,
  status,
  statusDescription,
  componentType,
  hasDrift,
  href,
}: ITopologyNodeProps) {
  const navigate = useNavigate()
  const theme = getStatusTheme(status)
  const isRunning = RUNNING_STATUSES.has(status)
  const needsApproval = APPROVAL_STATUSES.has(status)

  const typeConfig =
    variant === 'component' && componentType
      ? COMPONENT_TYPE_CONFIG[componentType as TComponentType] ?? COMPONENT_TYPE_CONFIG.unknown
      : null

  const iconVariant: TIconVariant = typeConfig?.icon ?? VARIANT_ICON[variant] ?? 'Cube'
  const typeLabel = typeConfig?.label ?? null

  const statusLabel = STATUS_LABEL_OVERRIDE[status] ?? toSentenceCase(status.replace(/[-_]/g, ' '))

  const dotClass = isRunning
    ? THEME_DOT.info
    : needsApproval
      ? THEME_DOT.warn
      : THEME_DOT[theme] ?? THEME_DOT.neutral

  const glowClass = needsApproval ? undefined : ({
    info:    'node-glow-running',
    error:   'node-glow-error',
    warn:    'node-glow-warn',
    success: 'node-glow-success',
  } as Record<string, string>)[theme]

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => navigate(href)}
        title={statusDescription || name}
        className={cn(
          'flex items-center gap-3 p-2 rounded-xl border text-left w-full cursor-pointer relative',
          'bg-white dark:bg-dark-grey-800',
          'border-cool-grey-200 dark:border-dark-grey-700',
          'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700',
          'hover:border-primary-300 dark:hover:border-primary-700',
          'transition-colors duration-150',
          needsApproval && 'approval-pulse',
          glowClass,
        )}
      >
      {/* Icon circle */}
      <div className="relative shrink-0 w-10 h-10 flex items-center justify-center">
        <div className={cn(
          'w-10 h-10 rounded-full flex items-center justify-center relative',
          'bg-cool-grey-200 dark:bg-dark-grey-700',
        )}>
        {isRunning ? (
          <svg
            className="h-5 w-5 animate-spin text-cool-grey-700 dark:text-cool-grey-400"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle className="opacity-25" cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="2" />
            <path className="opacity-75" stroke="currentColor" strokeWidth="2" strokeLinecap="round" d="M4 12a8 8 0 018-8" />
          </svg>
        ) : (
          <Icon
            variant={iconVariant}
            size={20}
            weight="bold"
            className="text-cool-grey-800 dark:text-cool-grey-300"
          />
        )}
        {/* Status dot on bottom-right corner */}
        <div className="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-dark-grey-800 flex items-center justify-center">
          <div className={cn('w-full h-full rounded-full', isRunning ? `${dotClass} animate-pulse` : dotClass)} />
        </div>
        </div>
      </div>

      {/* Text content */}
      <div className="flex flex-col min-w-0 flex-1 gap-0.5">
        <div className="flex items-center gap-1.5 min-w-0">
          <Text variant="body" weight="strong" className="truncate text-[13px] leading-snug">
            {name}
          </Text>
          {hasDrift && (
            <Icon variant="WarningCircle" size={12} weight="fill" className="text-orange-500 shrink-0" />
          )}
        </div>
        <Text
          variant="subtext"
          className={cn('truncate text-[12px]', isRunning ? THEME_TEXT.info : THEME_TEXT[theme] ?? THEME_TEXT.neutral)}
        >
          {statusLabel}
        </Text>
      </div>
    </button>
  </div>
  )
}

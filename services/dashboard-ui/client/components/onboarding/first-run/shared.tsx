import { useEffect, useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { CloudRegion } from '@/components/common/CloudRegion'
import { Code } from '@/components/common/Code'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { CLOUD_ICON, CLOUD_LABEL, TEST_CLOUDS, type TCloud } from './constants'

export const FirstRunCloudRegion = ({
  cloud,
  region,
  ...textProps
}: { cloud: TCloud; region: string } & Omit<IText, 'children'>) => (
  <CloudRegion
    platform={cloud}
    region={cloud === 'azure' ? undefined : region}
    location={cloud === 'azure' ? region : undefined}
    {...textProps}
  />
)

export const NextButton = ({
  label,
  disabled,
  disabledReason,
  loading,
  onClick,
  onBack,
  showNext = true,
  secondary,
  size = 'md',
}: {
  label?: string
  disabled?: boolean
  disabledReason?: string
  loading?: boolean
  onClick?: () => void
  onBack?: () => void
  showNext?: boolean
  secondary?: ReactNode
  size?: 'md' | 'lg'
}) => (
  <div className={cn('flex flex-wrap gap-3', onBack ? 'justify-between' : 'justify-end')}>
    {onBack ? (
      <Button type="button" variant="secondary" size={size} onClick={onBack}>
        <Icon variant="CaretLeftIcon" weight="bold" /> Back
      </Button>
    ) : null}
    <div className="flex flex-wrap items-center gap-3">
      {secondary}
      {showNext ? (
        <Button
          type="button"
          variant="primary"
          size={size}
          disabled={disabled || loading}
          onClick={onClick}
          tooltipProps={disabled && disabledReason ? { tipContent: disabledReason } : undefined}
        >
          {loading ? <Icon variant="Loading" size={16} /> : null}
          {label ?? 'Continue'}
          {loading ? null : <Icon variant="CaretRightIcon" weight="bold" />}
        </Button>
      ) : null}
    </div>
  </div>
)

export const CopyTextButton = ({
  text,
  label,
  size = 'md',
  variant = 'secondary',
  disabled,
  disabledReason,
  onCopy,
}: {
  text: string
  label: string
  size?: 'lg' | 'md' | 'sm'
  variant?: 'primary' | 'secondary'
  disabled?: boolean
  disabledReason?: string
  onCopy?: () => void
}) => {
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) return
    const timer = setTimeout(() => setCopied(false), 2000)
    return () => clearTimeout(timer)
  }, [copied])

  return (
    <Button
      variant={variant}
      size={size}
      className="shrink-0"
      disabled={disabled}
      tooltipProps={disabled && disabledReason ? { tipContent: disabledReason } : undefined}
      onClick={() => {
        navigator.clipboard?.writeText(text).catch(() => {})
        setCopied(true)
        onCopy?.()
      }}
    >
      <Icon variant={copied ? 'CheckIcon' : 'CopyIcon'} size={14} />
      {copied ? 'Copied' : label}
    </Button>
  )
}

export const NuonMark = ({ className }: { className?: string }) => (
  <svg viewBox="0 0 23.119 32" fill="none" className={className} role="img" aria-label="Nuon">
    <path
      fill="currentColor"
      fillRule="nonzero"
      d="M 16.994 0 L 10.87 3.537 L 10.87 9.263 L 5.912 6.398 L 5.91 6.398 L 0 9.811 L 0 28.588 L 5.907 32 L 5.91 32 L 12.251 28.336 L 12.251 22.862 L 16.994 25.599 L 23.119 22.062 L 23.119 3.537 L 16.994 0 Z M 1.384 10.61 L 5.907 8 L 5.91 8 L 10.867 10.862 L 10.867 20.463 L 1.384 14.989 L 1.384 10.61 Z M 10.867 27.537 L 5.907 30.398 L 1.384 27.788 L 1.384 16.588 L 10.867 22.062 L 10.867 27.537 L 10.867 27.537 Z M 21.734 21.26 L 16.994 23.997 L 12.254 21.263 L 12.254 11.661 L 21.737 17.136 L 21.737 21.26 L 21.734 21.26 Z M 21.734 15.537 L 12.251 10.062 L 12.251 4.336 L 16.994 1.599 L 21.734 4.336 L 21.734 15.537 Z"
    />
  </svg>
)

const APP_TIERS: { icon: TIconVariant; label: string }[] = [
  { icon: 'GlobeIcon', label: 'Web' },
  { icon: 'CubeIcon', label: 'API' },
  { icon: 'DatabaseIcon', label: 'Database' },
]

export const MiniArch = ({ live = false }: { live?: boolean }) => (
  <div className="flex flex-wrap items-center gap-y-2">
    {APP_TIERS.map((tier, index) => (
      <div key={tier.label} className="flex items-center">
        {index > 0 ? (
          <span className="h-px w-4 bg-neutral-200 dark:bg-neutral-600" aria-hidden />
        ) : null}
        <div className="flex items-center gap-1.5 rounded-md border bg-background px-2.5 py-1.5">
          <Icon variant={tier.icon} size={14} theme={live ? 'brand' : 'neutral'} />
          <Text variant="subtext" weight="strong">
            {tier.label}
          </Text>
        </div>
      </div>
    ))}
  </div>
)

export const InlineLink = ({
  href,
  children,
  textVariant = 'subtext',
}: {
  href: string
  children: ReactNode
  textVariant?: 'subtext' | 'body'
}) => (
  <Link href={href} isExternal textVariant={textVariant} className="!inline-flex align-baseline">
    {children}
  </Link>
)

export const RepoChip = ({ repo }: { repo: string }) => (
  <Link href={`https://github.com/${repo}`} isExternal textVariant="subtext">
    <Code variant="inline">{repo}</Code>
  </Link>
)

export const TestCloudPicker = ({
  value,
  onChange,
  error,
  disabled,
}: {
  value?: TCloud
  onChange: (cloud: TCloud) => void
  error: boolean
  disabled?: boolean
}) => (
  <fieldset aria-describedby={error ? 'test-cloud-error' : 'test-cloud-hint'}>
    <legend className="mb-2">
      <Text variant="body" weight="strong">
        Test cloud where your app will be installed
      </Text>
    </legend>
    <div className="flex flex-col gap-2">
      <div className="grid max-w-md grid-cols-3 gap-3">
        {TEST_CLOUDS.map((cloud) => {
          const checked = value === cloud
          return (
            <label
              key={cloud}
              title={CLOUD_LABEL[cloud]}
              className={cn(
                'flex h-14 items-center gap-3 rounded-md px-4 ring-1 transition-shadow',
                disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
                'has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-primary-500',
                checked
                  ? 'ring-2 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                  : error
                    ? 'ring-red-500 dark:ring-red-400 hover:bg-neutral-50 dark:hover:bg-neutral-900'
                    : 'ring-neutral-200 dark:ring-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-900'
              )}
            >
              <input
                type="radio"
                name="test-cloud"
                value={cloud}
                checked={checked}
                disabled={disabled}
                onChange={() => onChange(cloud)}
                aria-invalid={error || undefined}
                className="accent-primary-600 focus-visible:outline-none"
              />
              <span className="flex flex-1 justify-center">
                <Icon variant={CLOUD_ICON[cloud]} size={cloud === 'aws' ? 26 : 22} />
              </span>
              <span className="sr-only">{CLOUD_LABEL[cloud]}</span>
            </label>
          )
        })}
      </div>
      {error ? (
        <Text id="test-cloud-error" variant="subtext" theme="error" flex>
          <Icon variant="WarningCircleIcon" size={14} weight="fill" />
          Select a test cloud to continue.
        </Text>
      ) : (
        <Text id="test-cloud-hint" variant="subtext" theme="neutral">
          Nuon stubs the runner, sandbox and permissions for this cloud.
        </Text>
      )}
    </div>
  </fieldset>
)

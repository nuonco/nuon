'use client'

import { Icon } from '@/components/common/Icon'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { type ITooltip } from '@/components/common/Tooltip'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromInstallComponents,
} from '@/components/components/ComponentsTooltip'
import { useInstall } from '@/hooks/use-install'
import type { TInstall, TInstallComponent } from '@/types'
import { cn } from '@/utils/classnames'
import { getInstallStatusTitle } from '@/utils/install-utils'
import { toSentenceCase } from '@/utils/string-utils'

interface IInstallStatuses
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  isLabelHidden?: boolean
  tooltipPosition?: ITooltip['position']
  install: TInstall
}

type TStatusConfig = {
  label: string
  status: string
  statusDescription: string
  viewPath: string
}

const STATUS_CONFIGS: TStatusConfig[] = [
  {
    label: 'Runner',
    status: 'runner_status',
    statusDescription: 'runner_status_description',
    viewPath: 'runner',
  },
  {
    label: 'Sandbox',
    status: 'sandbox_status',
    statusDescription: 'sandbox_status_description',
    viewPath: 'sandbox',
  },
  /* {
   *   label: "Components",
   *   status: "composite_component_status",
   *   statusDescription: "composite_component_status_description",
   *   viewPath: "components",
   * }, */
]

function getTooltip({
  title,
  description,
  viewHref,
  viewLabel,
}: {
  title: string
  description: string
  viewHref: string
  viewLabel: string
}) {
  return (
    <div className="flex flex-col w-56">
      <Text className="leading-tight" weight="strong">
        {title}
      </Text>
      <Text variant="subtext">{description}</Text>
      <Text className="mt-2" variant="subtext">
        <Link className="flex items-center" href={viewHref}>
          View {viewLabel} <Icon variant="CaretRight" />
        </Link>
      </Text>
    </div>
  )
}

export const InstallStatuses = ({
  className,
  install,
  isLabelHidden = false,
  tooltipPosition = 'bottom',
  ...props
}: IInstallStatuses) => {
  const runnerContent = (
    <ContextTooltip
      title="Install runner"
      position={tooltipPosition}
      items={[
        {
          href: `/${install.org_id}/installs/${install.id}/runner`,
          id: install?.runner_id,
          title: `${install.runner_type === 'aws' ? 'AWS' : toSentenceCase(install?.runner_type)} runner`,
          subtitle: getInstallStatusTitle(
            'runner_status',
            install?.runner_status
          ),
          leftContent: (
            <Status
              status={install?.runner_status}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]}
    >
      <Status status={install.runner_status} variant="badge">
        {isLabelHidden ? 'Runner' : install.runner_status}
      </Status>
    </ContextTooltip>
  )

  const sandboxContent = (
    <ContextTooltip
      title="Latest sandbox run"
      position={tooltipPosition}
      items={[
        {
          href: `/${install.org_id}/installs/${install.id}/sandbox`,
          id: install?.install_sandbox_runs?.at(0)?.id,
          title: toSentenceCase(install?.install_sandbox_runs?.at(0)?.run_type),
          subtitle: getInstallStatusTitle(
            'sandbox_status',
            install?.sandbox_status
          ),
          leftContent: (
            <Status
              status={install.sandbox_status}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]}
    >
      <Status status={install.sandbox_status} variant="badge">
        {isLabelHidden ? 'Sandbox' : install.sandbox_status}
      </Status>
    </ContextTooltip>
  )

  const componentsContent = (
    <ComponentsTooltip
      title={getInstallStatusTitle(
        'composite_component_status',
        install?.composite_component_status
      )}
      componentSummaries={getContextTooltipItemsFromInstallComponents(
        install.install_components as TInstallComponent[],
        `/${install.org_id}/installs/${install.id}/components`
      )}
      position={tooltipPosition}
    >
      <Status status={install.composite_component_status} variant="badge">
        {isLabelHidden ? 'Components' : install.composite_component_status}
      </Status>
    </ComponentsTooltip>
  )

  return (
    <div className={cn('flex items-center gap-4', className)} {...props}>
      {isLabelHidden ? (
        runnerContent
      ) : (
        <LabeledValue label="Runner">{runnerContent}</LabeledValue>
      )}

      {isLabelHidden ? (
        sandboxContent
      ) : (
        <LabeledValue label="Sandbox">{sandboxContent}</LabeledValue>
      )}

      {isLabelHidden ? (
        componentsContent
      ) : (
        <LabeledValue label="Components">{componentsContent}</LabeledValue>
      )}
    </div>
  )
}

export const InstallStatusesContainer = (
  props: Omit<IInstallStatuses, 'install'>
) => {
  const { install } = useInstall()
  return <InstallStatuses install={install} {...props} />
}

export const SimpleInstallStatuses = ({
  install,
  isLabelHidden = false,
}: {
  install: TInstall
  isLabelHidden?: boolean
}) => {
  const RunnerStatus = (
    <Status status={install.runner_status} variant="badge">
      {isLabelHidden ? 'Runner' : install.runner_status}
    </Status>
  )
  const SandboxStatus = (
    <Status status={install.sandbox_status} variant="badge">
      {isLabelHidden ? 'Sandbox' : install.sandbox_status}
    </Status>
  )
  const ComponentsStatus = (
    <Status status={install.composite_component_status} variant="badge">
      {isLabelHidden ? 'Components' : install.composite_component_status}
    </Status>
  )

  return (
    <div className={cn('flex items-center gap-4')}>
      {isLabelHidden ? (
        RunnerStatus
      ) : (
        <LabeledValue label="Runner">{RunnerStatus}</LabeledValue>
      )}

      {isLabelHidden ? (
        SandboxStatus
      ) : (
        <LabeledValue label="Sandbox">{SandboxStatus}</LabeledValue>
      )}

      {isLabelHidden ? (
        ComponentsStatus
      ) : (
        <LabeledValue label="Components">{ComponentsStatus}</LabeledValue>
      )}
    </div>
  )
}

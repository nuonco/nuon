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
import type { TInstall, TInstallComponent, TInstallStack } from '@/types'
import { cn } from '@/utils/classnames'
import { Time } from '@/components/common/Time'
import { getInstallStatusTitle } from '@/utils/install-utils'
import { getStatusTheme, getWorstStatusTheme } from '@/utils/status-utils'
import { toSentenceCase } from '@/utils/string-utils'

export interface IInstallStatuses
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  collapsible?: boolean
  isLabelHidden?: boolean
  tooltipPosition?: ITooltip['position']
  variant?: 'badge' | 'icon'
  install: TInstall
  stack?: TInstallStack
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
          View {viewLabel} <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    </div>
  )
}

export const InstallStatuses = ({
  className,
  collapsible = false,
  install,
  isLabelHidden = false,
  tooltipPosition = 'bottom',
  variant = 'badge',
  stack,
  ...props
}: IInstallStatuses) => {
  const lifecycleStatus = install?.lifecycle_phase?.phase
  const isDeprovisioning = lifecycleStatus === 'deprovisioning'
  const isDeprovisioned = lifecycleStatus === 'deprovisioned'

  const STALE_STATUSES = ['active', 'pending', 'executing', 'queued', 'planning', 'syncing']
  const effectiveStatus = (status: string | undefined) => {
    if (!status) return status
    if (isDeprovisioned && STALE_STATUSES.includes(status)) return 'deprovisioned'
    if (isDeprovisioning && STALE_STATUSES.includes(status)) return 'deprovisioning'
    return status
  }

  const driftStatus = install?.drifted_objects?.length ? 'warn' : 'active'
  const latestStackVersion = stack?.versions?.[0]
  const stackStatus = latestStackVersion?.composite_status?.status

  const driftContent = (
    <ContextTooltip
      title="Drift detection"
      position={tooltipPosition}
      items={
        install?.drifted_objects?.length
          ? install?.drifted_objects?.map((drift) => ({
              href: `/${install.org_id}/installs/${install.id}/workflows/${drift?.install_workflow_id}`,
              id: drift?.target_id,
              title:
                drift?.target_type === 'install_deploy'
                  ? drift?.component_name
                  : 'Sandbox',
              subtitle: 'Drift detected',
              leftContent: (
                <Status
                  status="warn"
                  isWithoutText
                  variant="timeline"
                  iconSize={16}
                />
              ),
            }))
          : [
              {
                id: install?.runner_id,
                title: 'No drift',
                subtitle: 'Install has detected no drift',
                leftContent: (
                  <Status
                    status={install?.runner_status}
                    isWithoutText
                    variant="timeline"
                    iconSize={16}
                  />
                ),
              },
            ]
      }
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(driftStatus)}>
          <Icon variant="FileDashedIcon" size={14} className="cursor-default" />
        </Text>
      ) : (
        <Status
          status={driftStatus}
          variant="badge"
          className={
            install.drifted_objects?.length
              ? '[&>span:first-child]:animate-pulse'
              : undefined
          }
        >
          {isLabelHidden
            ? 'Drift'
            : install.drifted_objects?.length
              ? 'Drifted'
              : 'No drift'}
        </Status>
      )}
    </ContextTooltip>
  )

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
            install?.runner_status,
            install?.lifecycle_phase?.phase
          ),
          leftContent: (
            <Status
              status={effectiveStatus(install?.runner_status)}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]}
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(effectiveStatus(install.runner_status) ?? '')}>
          <Icon
            variant="SneakerMoveIcon"
            size={14}
            className="cursor-default"
          />
        </Text>
      ) : (
        <Status status={effectiveStatus(install.runner_status)} variant="badge">
          {isLabelHidden ? 'Runner' : effectiveStatus(install.runner_status)}
        </Status>
      )}
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
            install?.sandbox_status,
            install?.lifecycle_phase?.phase
          ),
          leftContent: (
            <Status
              status={effectiveStatus(install.sandbox_status)}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]}
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(effectiveStatus(install.sandbox_status) ?? '')}>
          <Icon
            variant="ShippingContainerIcon"
            size={14}
            className="cursor-default"
          />
        </Text>
      ) : (
        <Status status={effectiveStatus(install.sandbox_status)} variant="badge">
          {isLabelHidden ? 'Sandbox' : effectiveStatus(install.sandbox_status)}
        </Status>
      )}
    </ContextTooltip>
  )

  const componentsContent = (
    <ComponentsTooltip
      title={getInstallStatusTitle(
        'composite_component_status',
        install?.composite_component_status,
        install?.lifecycle_phase?.phase
      )}
      componentSummaries={getContextTooltipItemsFromInstallComponents(
        install?.install_components as TInstallComponent[],
        `/${install.org_id}/installs/${install.id}/components`,
        install?.lifecycle_phase?.phase
      )}
      position={tooltipPosition}
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(effectiveStatus(install.composite_component_status) ?? '')}>
          <Icon variant="CardsIcon" size={14} className="cursor-default" />
        </Text>
      ) : (
        <Status status={effectiveStatus(install.composite_component_status)} variant="badge">
          {isLabelHidden ? 'Components' : effectiveStatus(install.composite_component_status)}
        </Status>
      )}
    </ComponentsTooltip>
  )

  const stackContent = stackStatus ? (
    <ContextTooltip
      title="Stack"
      position={tooltipPosition}
      items={[
        {
          href: `/${install.org_id}/installs/${install.id}/stacks`,
          id: latestStackVersion?.id ?? 'stack',
          title: toSentenceCase(stackStatus),
          subtitle: latestStackVersion?.created_at ? (
            <Time
              time={latestStackVersion.created_at}
              variant="label"
              theme="neutral"
            />
          ) : (
            latestStackVersion?.composite_status?.status_human_description
          ),
          leftContent: (
            <Status
              status={stackStatus}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
      ]}
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(stackStatus)}>
          <Icon variant="StackIcon" size={14} className="cursor-default" />
        </Text>
      ) : (
        <Status status={stackStatus} variant="badge">
          {isLabelHidden ? 'Stack' : stackStatus}
        </Status>
      )}
    </ContextTooltip>
  ) : null

  const isIcon = variant === 'icon'

  const wrap = (label: string, content: React.ReactNode) =>
    isIcon || isLabelHidden ? (
      content
    ) : (
      <LabeledValue label={label}>{content}</LabeledValue>
    )

  const allStatuses = [
    install?.drifted_objects?.length ? 'warn' : 'active',
    stackStatus,
    effectiveStatus(install?.runner_status),
    effectiveStatus(install?.sandbox_status),
    effectiveStatus(install?.composite_component_status),
  ]
  const { worstStatus } = getWorstStatusTheme(allStatuses)

  const compositeItems = [
    ...(install?.drifted_objects
      ? [
          {
            id: 'drift',
            title: 'Drift detection',
            subtitle: install.drifted_objects.length
              ? 'Drift detected'
              : 'No drift',
            href: install.drifted_objects.length
              ? `/${install.org_id}/installs/${install.id}/workflows`
              : undefined,
            leftContent: (
              <Status
                status={install.drifted_objects.length ? 'warn' : 'active'}
                isWithoutText
                variant="timeline"
                iconSize={16}
              />
            ),
          },
        ]
      : []),
    ...(stackStatus
      ? [
          {
            id: 'stack',
            title: 'Stack',
            subtitle: toSentenceCase(stackStatus),
            href: `/${install.org_id}/installs/${install.id}/stacks`,
            leftContent: (
              <Status
                status={stackStatus}
                isWithoutText
                variant="timeline"
                iconSize={16}
              />
            ),
          },
        ]
      : []),
    {
      id: 'runner',
      title: 'Runner',
      subtitle: getInstallStatusTitle('runner_status', install?.runner_status, install?.lifecycle_phase?.phase),
      href: `/${install.org_id}/installs/${install.id}/runner`,
      leftContent: (
        <Status
          status={effectiveStatus(install?.runner_status)}
          isWithoutText
          variant="timeline"
          iconSize={16}
        />
      ),
    },
    {
      id: 'sandbox',
      title: 'Sandbox',
      subtitle: getInstallStatusTitle(
        'sandbox_status',
        install?.sandbox_status,
        install?.lifecycle_phase?.phase
      ),
      href: `/${install.org_id}/installs/${install.id}/sandbox`,
      leftContent: (
        <Status
          status={effectiveStatus(install?.sandbox_status)}
          isWithoutText
          variant="timeline"
          iconSize={16}
        />
      ),
    },
    {
      id: 'components',
      title: 'Components',
      subtitle: getInstallStatusTitle(
        'composite_component_status',
        install?.composite_component_status,
        install?.lifecycle_phase?.phase
      ),
      href: `/${install.org_id}/installs/${install.id}/components`,
      leftContent: (
        <Status
          status={effectiveStatus(install?.composite_component_status)}
          isWithoutText
          variant="timeline"
          iconSize={16}
        />
      ),
    },
  ]

  const expandedContent = (
    <div className={cn('flex items-center flex-wrap gap-2', { 'hidden @5xl:flex': collapsible })}>
      {install?.drifted_objects ? wrap('Drift detection', driftContent) : null}
      {stackContent ? wrap('Stack', stackContent) : null}
      {wrap('Runner', runnerContent)}
      {wrap('Sandbox', sandboxContent)}
      {wrap('Components', componentsContent)}
    </div>
  )

  const collapsedContent = collapsible ? (
    <div className="flex @5xl:hidden">
      <LabeledValue label="Status">
        <ContextTooltip
          title="Install status"
          items={compositeItems}
          position={tooltipPosition}
        >
          <Status status={worstStatus} variant="badge">
            {worstStatus}
          </Status>
        </ContextTooltip>
      </LabeledValue>
    </div>
  ) : null

  return (
    <div className={cn(className)} {...props}>
      {expandedContent}
      {collapsedContent}
    </div>
  )
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

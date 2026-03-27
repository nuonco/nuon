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
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TInstall, TInstallComponent, TInstallStack } from '@/types'
import { cn } from '@/utils/classnames'
import { Time } from '@/components/common/Time'
import { getInstallStatusTitle } from '@/utils/install-utils'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase } from '@/utils/string-utils'

interface IInstallStatuses
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
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
  variant = 'badge',
  stack,
  ...props
}: IInstallStatuses) => {
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
        <Status status={driftStatus} variant="badge">
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
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(install.runner_status ?? '')}>
          <Icon
            variant="SneakerMoveIcon"
            size={14}
            className="cursor-default"
          />
        </Text>
      ) : (
        <Status status={install.runner_status} variant="badge">
          {isLabelHidden ? 'Runner' : install.runner_status}
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
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(install.sandbox_status ?? '')}>
          <Icon
            variant="ShippingContainerIcon"
            size={14}
            className="cursor-default"
          />
        </Text>
      ) : (
        <Status status={install.sandbox_status} variant="badge">
          {isLabelHidden ? 'Sandbox' : install.sandbox_status}
        </Status>
      )}
    </ContextTooltip>
  )

  const componentsContent = (
    <ComponentsTooltip
      title={getInstallStatusTitle(
        'composite_component_status',
        install?.composite_component_status
      )}
      componentSummaries={getContextTooltipItemsFromInstallComponents(
        install?.install_components as TInstallComponent[],
        `/${install.org_id}/installs/${install.id}/components`
      )}
      position={tooltipPosition}
    >
      {variant === 'icon' ? (
        <Text theme={getStatusTheme(install.composite_component_status ?? '')}>
          <Icon variant="CardsIcon" size={14} className="cursor-default" />
        </Text>
      ) : (
        <Status status={install.composite_component_status} variant="badge">
          {isLabelHidden ? 'Components' : install.composite_component_status}
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
          <Icon variant="Stack" size={14} className="cursor-default" />
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

  return (
    <div className={cn('flex items-center flex-wrap gap-2', className)} {...props}>
      {install?.drifted_objects ? wrap('Drift detection', driftContent) : null}
      {stackContent ? wrap('Stack', stackContent) : null}
      {wrap('Runner', runnerContent)}
      {wrap('Sandbox', sandboxContent)}
      {wrap('Components', componentsContent)}
    </div>
  )
}

export const InstallStatusesContainer = (
  props: Omit<IInstallStatuses, 'install' | 'stack'>
) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: stack } = useQuery({
    queryKey: ['install-stack', org.id, install.id],
    queryFn: () => getInstallStack({ installId: install.id, orgId: org.id }),
    enabled: !!install.id,
  })
  return <InstallStatuses install={install} stack={stack} {...props} />
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

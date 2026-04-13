import React, { memo, useCallback } from 'react'
import { useNavigate } from 'react-router'
import type { NodeProps } from '@xyflow/react'
import { Icon } from '@/components/common/Icon'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import type { TIconVariant } from '@/components/common/Icon'
import type { TComponentType } from '@/types'
import { cn } from '@/utils/classnames'
import { toSentenceCase } from '@/utils/string-utils'
import type { TRoleInfo } from './diagram-layout'

const CONTAINER_STYLES: Record<number, string> = {
  0: 'border border-cool-grey-400 bg-cool-grey-200 dark:border-dark-grey-300 dark:bg-dark-grey-800',
  1: 'border border-dashed border-cool-grey-400 bg-cool-grey-100 dark:border-dark-grey-200 dark:bg-dark-grey-700',
  2: 'border border-cool-grey-300 bg-cool-grey-50 dark:border-dark-grey-200 dark:bg-dark-grey-900',
  3: 'border border-cool-grey-300 bg-white dark:border-dark-grey-100 dark:bg-dark-grey-800',
}

export const ContainerNode = memo(({ data }: NodeProps) => {
  const navigate = useNavigate()
  const level = (data.level as number) ?? 0
  const status = data.status as string
  const width = data.width as number
  const height = data.height as number
  const href = data.href as string | undefined

  const handleClick = useCallback(() => {
    if (href) navigate(href)
  }, [href, navigate])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (href && (e.key === 'Enter' || e.key === ' ')) {
        e.preventDefault()
        navigate(href)
      }
    },
    [href, navigate]
  )

  return (
    <div
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role={href ? 'button' : undefined}
      tabIndex={href ? 0 : undefined}
      className={cn(
        'rounded-lg nopan nodrag',
        CONTAINER_STYLES[level] || 'border bg-white dark:bg-dark-grey-900',
        href && 'cursor-pointer'
      )}
      style={{ width, height }}
    >
      <div className="flex items-center gap-1.5 px-3 py-2">
        <Icon
          variant={data.icon as TIconVariant}
          size={14}
          theme="neutral"
        />
        <Text variant="label" weight="strong" theme="neutral">
          {data.label as string}
        </Text>
        {status ? <Status status={status} variant="badge" /> : null}
      </div>
    </div>
  )
})

ContainerNode.displayName = 'ContainerNode'

export const SectionLabelNode = memo(({ data }: NodeProps) => (
  <div className="flex items-center gap-1.5">
    <Icon variant={data.icon as TIconVariant} size={12} theme="neutral" />
    <Text variant="label" weight="strong" theme="neutral">
      {data.label as string}
    </Text>
  </div>
))

SectionLabelNode.displayName = 'SectionLabelNode'

const COMPONENT_TYPE_LABELS: Record<string, string> = {
  helm_chart: 'Helm chart',
  terraform_module: 'Terraform module',
  kubernetes_manifest: 'Kubernetes manifest',
  docker_build: 'Docker build',
  external_image: 'External image',
  job: 'Job',
  pulumi: 'Pulumi',
}

export const ComponentCardNode = memo(({ data }: NodeProps) => {
  const status = data.status as string
  const isDrifted = data.isDrifted as boolean
  const width = data.width as number
  const componentType = data.componentType as TComponentType
  const href = data.href as string
  const name = data.name as string
  const latestDeployAt = data.latestDeployAt as string | undefined
  const deploysHref = data.deploysHref as string | undefined

  return (
    <ContextTooltip
      className="nopan nodrag"
      position="bottom"
      width="w-64"
      title={name}
      items={[
        {
          id: 'status',
          title: 'Status',
          subtitle:
            (data.statusDescription as string) || toSentenceCase(status),
          leftContent: (
            <Status
              status={status}
              isWithoutText
              variant="timeline"
              iconSize={16}
            />
          ),
        },
        ...(data.latestDeployStatus
          ? [
              {
                id: 'deploy',
                title: 'Latest deploy',
                subtitle: (
                  <span className="flex items-center gap-1">
                    <Text variant="label" theme="neutral">
                      {toSentenceCase(data.latestDeployStatus as string)}
                    </Text>
                    {latestDeployAt && (
                      <>
                        <Text variant="label" theme="neutral">·</Text>
                        <Time
                          time={latestDeployAt}
                          format="relative"
                          variant="label"
                          theme="neutral"
                        />
                      </>
                    )}
                  </span>
                ),
                leftContent: (
                  <Status
                    status={data.latestDeployStatus as string}
                    isWithoutText
                    variant="timeline"
                    iconSize={16}
                  />
                ),
              },
            ]
          : []),
        ...(isDrifted
          ? [
              {
                id: 'drift',
                title: 'Drift detected',
                subtitle: 'Configuration has drifted from desired state',
                leftContent: (
                  <Icon variant="WarningIcon" size={16} theme="warn" />
                ),
              },
            ]
          : []),
        {
          id: 'view',
          title: 'View details',
          subtitle: COMPONENT_TYPE_LABELS[componentType] || 'Component',
          href,
          leftContent: (
            <ComponentType
              type={componentType}
              displayVariant="icon-only"
              variant="label"
              colorVariant="color"
              iconSize="16"
            />
          ),
        },
        ...(deploysHref
          ? [
              {
                id: 'deploys',
                title: 'View deploys',
                subtitle: 'Deploy history & logs',
                href: deploysHref,
                leftContent: (
                  <Icon variant="ClockCounterClockwise" size={16} />
                ),
              },
            ]
          : []),
      ]}
    >
      <div
        aria-label={`${name} — ${COMPONENT_TYPE_LABELS[componentType] || 'Component'}`}
        className={cn(
          'rounded-lg border bg-white dark:bg-dark-grey-900 px-3 py-2 flex items-center gap-2 transition-shadow hover:shadow-sm',
          isDrifted && '!border-orange-400 dark:!border-orange-500/40'
        )}
        style={{ width, height: 56 }}
      >
        <ComponentType
          type={componentType}
          displayVariant="icon-only"
          variant="label"
          colorVariant="color"
          iconSize="16"
        />
        <div className="flex flex-col gap-0.5 min-w-0 flex-1">
          <Text
            variant="label"
            weight="strong"
            className="truncate !block"
          >
            {name}
          </Text>
          <Text variant="label" theme="neutral" className="truncate !block">
            {COMPONENT_TYPE_LABELS[componentType] || 'Component'}
          </Text>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {isDrifted && (
            <Icon variant="WarningIcon" size={12} theme="warn" />
          )}
          <Status status={status || 'not-deployed'} variant="badge">
            {status ? undefined : 'Not deployed'}
          </Status>
        </div>
      </div>
    </ContextTooltip>
  )
})

ComponentCardNode.displayName = 'ComponentCardNode'

export const RoleCardNode = memo(({ data }: NodeProps) => {
  const role = data as unknown as TRoleInfo & { width: number }

  const items: import('@/components/common/ContextTooltip').TContextTooltipItem[] = [
    {
      id: 'status',
      title: 'Provisioned',
      subtitle: role.enabled ? 'Provisioned' : 'Not provisioned',
      leftContent: (
        <Status
          status={role.enabled ? 'active' : 'inactive'}
          isWithoutText
          variant="timeline"
          iconSize={16}
        />
      ),
    },
    ...(role.description
      ? [
          {
            id: 'description',
            title: 'Description',
            subtitle: role.description,
            leftContent: <Icon variant="Info" size={16} />,
          },
        ]
      : []),
    ...(role.policies.length > 0
      ? [
          {
            id: 'policies',
            title: `${role.policies.length} ${role.policies.length === 1 ? 'Policy' : 'Policies'}`,
            subtitle: (
              <span className="flex flex-col">
                {role.policies.slice(0, 3).map((p, i) => (
                  <Text key={i} variant="label" theme="neutral" family="mono" className="truncate">
                    {p.name || 'Unnamed'}
                  </Text>
                ))}
                {role.policies.length > 3 && (
                  <Text variant="label" theme="neutral">
                    +{role.policies.length - 3} more
                  </Text>
                )}
              </span>
            ),
            leftContent: <Icon variant="ShieldCheck" size={16} />,
          },
        ]
      : []),
  ]

  return (
    <ContextTooltip
      className="nopan nodrag"
      position="right"
      width="w-64"
      title={role.displayName}
      items={items}
    >
      <div
        className="rounded-lg border bg-white dark:bg-dark-grey-900 px-3 py-2 flex items-center justify-between cursor-default overflow-hidden"
        style={{ width: role.width, height: 48 }}
      >
        <Text variant="label" weight="strong" className="truncate">
          {role.name}
        </Text>
        <Status
          status={role.enabled ? 'active' : 'inactive'}
          variant="badge"
        />
      </div>
    </ContextTooltip>
  )
})

RoleCardNode.displayName = 'RoleCardNode'

export const nodeTypes = {
  containerNode: ContainerNode,
  sectionLabelNode: SectionLabelNode,
  componentCardNode: ComponentCardNode,
  roleCardNode: RoleCardNode,
}

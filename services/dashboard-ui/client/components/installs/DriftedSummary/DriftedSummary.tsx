import type { HTMLAttributes } from 'react'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TDriftedObject } from '@/types/ctl-api.types'
import { cn } from '@/utils/classnames'
import { useInstallLink } from '@/hooks/use-install-path'

export interface IDriftedSummary
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  driftedObjects: TDriftedObject[]
  orgId: string
  installId: string
}

export const DriftedSummary = ({
  driftedObjects,
  orgId,
  installId,
  className,
  ...props
}: IDriftedSummary) => {
  const installLink = useInstallLink()
  if (!driftedObjects.length) return null

  return (
    <div className={cn('flex items-center', className)} {...props}>
      <ContextTooltip
        title="Drift detected"
        position="bottom"
        showCount
        maxHeight="max-h-64"
        width="w-64"
        items={driftedObjects.map((drift) => ({
          id: drift?.target_id ?? '',
          href: installLink({ orgId: orgId, installId: installId, suffix: `/workflows/${drift?.install_workflow_id}` }),
          title:
            drift?.target_type === 'install_deploy'
              ? (drift?.component_name ?? 'Component')
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
        }))}
      >
        <Link
          variant="ghost"
          href={installLink({ orgId: orgId, installId: installId, suffix: `/workflows?type=drift_run` })}
        >
          <Text
            theme="warn"
            weight="strong"
            className="inline-flex items-center gap-1.5"
          >
            <Icon variant="WarningIcon" weight="bold" size={16} />
            Drift detected
          </Text>
          <Icon variant="CaretDownIcon" size={12} />
        </Link>
      </ContextTooltip>
    </div>
  )
}

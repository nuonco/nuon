import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'

export type TRemovedResourceKind = 'component' | 'action' | 'runbook'

const actionVerb: Record<TRemovedResourceKind, string> = {
  component: 'deployed',
  action: 'run',
  runbook: 'run',
}

export const RemovedFromAppConfigBadge = ({
  kind,
}: {
  kind: TRemovedResourceKind
}) => (
  <Tooltip
    position="top"
    tipContent={
      <Text variant="subtext">
        {`This ${kind} is no longer in the install's app config version and can't be ${actionVerb[kind]}.`}
      </Text>
    }
  >
    <Badge size="sm" theme="warn">
      Removed
    </Badge>
  </Tooltip>
)

export const RemovedFromAppConfigBanner = ({
  kind,
}: {
  kind: TRemovedResourceKind
}) => (
  <Banner theme="warn">
    <div className="flex flex-col gap-0.5">
      <Text weight="strong">Removed from the app config</Text>
      <Text variant="subtext">
        This {kind} is no longer in the install's app config version, so it can't
        be {actionVerb[kind]}. It's shown here for history.
      </Text>
    </div>
  </Banner>
)

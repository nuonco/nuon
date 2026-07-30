import { Code } from '@/components/common/Code'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TTriggerRule } from '@/types'

export const TriggerRuleDetails = ({
  orgId,
  rule,
}: {
  orgId: string
  rule: TTriggerRule
}) => {
  const appHref = rule?.app_id ? `/${orgId}/apps/${rule.app_id}` : undefined
  const appConfigHref =
    rule?.app_id && rule?.app_branch_id
      ? `/${orgId}/apps/${rule.app_id}/branches/${rule.app_branch_id}/configs`
      : appHref
  const isRunbookTarget = rule?.target_type === 'runbook' || !!rule?.runbook_id

  return (
    <div className="flex flex-col gap-6">
      <div className="grid gap-6 md:grid-cols-2">
        <LabeledValue label="App">
          {appHref ? (
            <Link href={appHref}>{rule?.app_name || rule.app_id}</Link>
          ) : (
            <Text variant="subtext">Unknown app</Text>
          )}
        </LabeledValue>
        <LabeledValue label="App config">
          {appConfigHref && rule?.app_config_id ? (
            <Link href={appConfigHref} className="font-mono">
              {rule.app_config_id}
            </Link>
          ) : (
            <Text variant="subtext">Unknown config</Text>
          )}
        </LabeledValue>
        <LabeledValue label="Event types">
          <Text variant="subtext">
            {rule?.event_types?.join(', ') || 'All event types'}
          </Text>
        </LabeledValue>
        <LabeledValue label="Target">
          {isRunbookTarget ? (
            <div className="flex flex-col gap-1">
              {appHref && rule?.runbook_id ? (
                <Link href={`${appHref}/runbooks/${rule.runbook_id}`}>
                  {rule?.runbook_name || rule.runbook_id}
                </Link>
              ) : (
                <Text variant="subtext">
                  {rule?.runbook_name || 'Unknown runbook'}
                </Text>
              )}
              <Text variant="subtext" theme="neutral">
                Install {rule?.install_name || 'not configured'}
              </Text>
            </div>
          ) : rule?.app_id && rule?.app_branch_id ? (
            <Link
              href={`/${orgId}/apps/${rule.app_id}/branches/${rule.app_branch_id}`}
            >
              {rule?.app_branch_name || rule.app_branch_id}
            </Link>
          ) : (
            <Text variant="subtext" theme="neutral">
              Unknown app branch
            </Text>
          )}
        </LabeledValue>
        <LabeledValue label="Status">
          <Status
            status={rule?.enabled === false ? 'neutral' : 'success'}
            variant="badge"
          >
            {rule?.enabled === false ? 'Disabled' : 'Enabled'}
          </Status>
        </LabeledValue>
      </div>
      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Filters
        </Text>
        {rule?.filters?.length ? (
          rule.filters.map((filter, index) => (
            <Code
              key={`${filter?.path}-${index}`}
              variant="preformated"
            >{`${filter?.from || 'payload'} ${filter?.path || '—'} ${filter?.op || '—'} ${JSON.stringify(filter?.value)}`}</Code>
          ))
        ) : (
          <Text variant="subtext" theme="neutral">
            No filters. Every configured event type matches.
          </Text>
        )}
      </div>
      {isRunbookTarget ? (
        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Input mappings
          </Text>
          {Object.keys(rule?.input_mappings ?? {}).length ? (
            Object.entries(rule?.input_mappings ?? {}).map(([input, path]) => (
              <Code
                key={input}
                variant="preformated"
              >{`${input} ← ${path}`}</Code>
            ))
          ) : (
            <Text variant="subtext" theme="neutral">
              No runbook inputs are mapped from the event.
            </Text>
          )}
        </div>
      ) : null}
    </div>
  )
}

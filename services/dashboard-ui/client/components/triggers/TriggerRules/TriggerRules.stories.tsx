export default { title: 'Triggers/Trigger rules' }
import { TriggerRules } from './TriggerRules'
const rules = [
  {
    id: 'rule-1',
    name: 'Deploy main',
    app_id: 'app-1',
    app_name: 'Storefront',
    app_config_id: 'cfg-1',
    app_branch_id: 'branch-1',
    app_branch_name: 'main',
    event_types: ['push'],
  },
]
export const Default = () => (
  <TriggerRules
    data={rules}
    hasError={false}
    isLoading={false}
    onRetry={() => {}}
    orgId="org-1"
    triggerId="trigger-1"
  />
)
export const Empty = () => (
  <TriggerRules
    data={[]}
    hasError={false}
    isLoading={false}
    onRetry={() => {}}
    orgId="org-1"
    triggerId="trigger-1"
  />
)
export const Error = () => (
  <TriggerRules
    data={[]}
    hasError
    isLoading={false}
    onRetry={() => {}}
    orgId="org-1"
    triggerId="trigger-1"
  />
)

export default { title: 'Triggers/Trigger rule details' }
import { TriggerRuleDetails } from './TriggerRuleDetails'
export const Default = () => (
  <TriggerRuleDetails
    orgId="org-1"
    rule={{
      id: 'rule-1',
      app_name: 'Storefront',
      app_config_id: 'cfg-1',
      app_id: 'app-1',
      app_branch_id: 'branch-1',
      app_branch_name: 'main',
      event_types: ['push'],
      filters: [{ path: '$.ref', op: 'eq', value: 'main' }],
    }}
  />
)

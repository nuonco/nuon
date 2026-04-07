export default {
  title: 'Actions/ActionTriggerType',
}

import { ActionTriggerType } from './ActionTriggerType'

export const Manual = () => <ActionTriggerType triggerType="manual" />

export const Cron = () => (
  <ActionTriggerType triggerType="cron" cronSchedule="0 9 * * 1-5" />
)

export const PreDeployComponent = () => (
  <ActionTriggerType
    triggerType="pre-deploy-component"
    componentName="api-server"
    componentPath="/org-1/installs/install-1/components/comp-1"
  />
)

export const PostDeployAll = () => (
  <ActionTriggerType triggerType="post-deploy-all-components" />
)

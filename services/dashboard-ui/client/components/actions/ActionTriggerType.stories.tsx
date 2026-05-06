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

export const PostDeployComponent = () => (
  <ActionTriggerType
    triggerType="post-deploy-component"
    componentName="rds_cluster_coder"
    componentPath="/org-1/installs/install-1/components/comp-1"
  />
)

export const PreTeardownComponent = () => (
  <ActionTriggerType
    triggerType="pre-teardown-component"
    componentName="api-server"
    componentPath="/org-1/installs/install-1/components/comp-1"
  />
)

export const PostTeardownComponent = () => (
  <ActionTriggerType
    triggerType="post-teardown-component"
    componentName="api-server"
    componentPath="/org-1/installs/install-1/components/comp-1"
  />
)

export const PostDeployAll = () => (
  <ActionTriggerType triggerType="post-deploy-all-components" />
)

export const PreDeployAll = () => (
  <ActionTriggerType triggerType="pre-deploy-all-components" />
)

export const ConstrainedWidth = () => (
  <div className="w-48 border p-2">
    <ActionTriggerType
      triggerType="post-deploy-component"
      componentName="really-long-component-name-that-overflows"
      componentPath="/org-1/installs/install-1/components/comp-1"
    />
  </div>
)

export default {
  title: 'Installs/RemovedFromAppConfig',
}

import {
  RemovedFromAppConfigBadge,
  RemovedFromAppConfigBanner,
} from './RemovedFromAppConfig'

export const Badges = () => (
  <div className="flex gap-2">
    <RemovedFromAppConfigBadge kind="component" />
    <RemovedFromAppConfigBadge kind="action" />
    <RemovedFromAppConfigBadge kind="runbook" />
  </div>
)

export const Banners = () => (
  <div className="flex flex-col gap-4">
    <RemovedFromAppConfigBanner kind="component" />
    <RemovedFromAppConfigBanner kind="action" />
    <RemovedFromAppConfigBanner kind="runbook" />
  </div>
)

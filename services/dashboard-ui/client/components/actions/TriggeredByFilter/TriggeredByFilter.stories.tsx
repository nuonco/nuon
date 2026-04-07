export default {
  title: 'Actions/TriggeredByFilter',
}

import { TriggeredByFilter } from './TriggeredByFilter'

const noop = () => {}

export const Default = () => (
  <TriggeredByFilter
    triggerType=""
    onTriggerChange={noop}
    onClearFilter={noop}
  />
)

export const WithSelection = () => (
  <TriggeredByFilter
    triggerType="manual"
    onTriggerChange={noop}
    onClearFilter={noop}
  />
)

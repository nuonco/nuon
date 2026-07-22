export default { title: 'Triggers/Trigger filters' }

import { TriggerFilters } from './TriggerFilters'

const noop = () => {}

export const Default = () => (
  <TriggerFilters
    trigger=""
    authType=""
    envelope=""
    onSourceChange={noop}
    onAuthTypeChange={noop}
    onEnvelopeChange={noop}
    onClearSource={noop}
    onClearAuthType={noop}
    onClearEnvelope={noop}
  />
)

export const WithSelections = () => (
  <TriggerFilters
    trigger="AWS"
    authType="hmac"
    envelope="cloudevents"
    onSourceChange={noop}
    onAuthTypeChange={noop}
    onEnvelopeChange={noop}
    onClearSource={noop}
    onClearAuthType={noop}
    onClearEnvelope={noop}
  />
)

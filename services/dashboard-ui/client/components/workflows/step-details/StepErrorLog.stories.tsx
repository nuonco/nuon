export default {
  title: 'Workflows/StepErrorLog',
}

import { StepErrorLog } from './StepErrorLog'

const renderErrorAttribute = [
  'unable to render interface string map value: unable to execute template:',
  'template: input:1:11: executing "input" at',
  '<.nuon.actions.workflows.dns_delegation.outputs.delegated>: map has no entry for key "delegated"',
].join(' ')

export const Default = () => <StepErrorLog text={renderErrorAttribute} />

export const WithLogsLink = () => (
  <StepErrorLog
    text={renderErrorAttribute}
    logsHref="/org-mock-001/installs/inst-mock-001/logs"
  />
)

export const ShortMessage = () => <StepErrorLog text="error rendering vars" />

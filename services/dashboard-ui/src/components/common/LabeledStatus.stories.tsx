import { LabeledStatus } from './LabeledStatus'

export const Default = () => (
  <LabeledStatus
    label="Status"
    statusProps={{ status: 'success' }}
    tooltipProps={{ tipContent: 'This is a status tooltip' }}
  />
)

export const WithDifferentStatuses = () => (
  <div className="flex gap-4 items-center">
    <LabeledStatus
      label="Success"
      statusProps={{ status: 'success' }}
      tooltipProps={{ tipContent: 'This is a success tooltip' }}
    />
    <LabeledStatus
      label="Failure"
      statusProps={{ status: 'failed' }}
      tooltipProps={{ tipContent: 'This is a failure tooltip' }}
    />
    <LabeledStatus
      label="Running"
      statusProps={{ status: 'warn' }}
      tooltipProps={{ tipContent: 'This is a running tooltip' }}
    />
    <LabeledStatus
      label="Queued"
      statusProps={{ status: 'queued' }}
      tooltipProps={{ tipContent: 'This is a queued tooltip' }}
    />
  </div>
)

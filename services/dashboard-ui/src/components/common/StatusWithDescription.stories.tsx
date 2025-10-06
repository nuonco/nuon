import { StatusWithDescription } from './StatusWithDescription'

export const Basic = () => (
  <div className="flex flex-col gap-4">
    <StatusWithDescription
      statusProps={{ status: 'success' }}
      tooltipProps={{ tipContent: 'Operation completed successfully' }}
    />
    <StatusWithDescription
      statusProps={{ status: 'error', variant: 'badge' }}
      tooltipProps={{ tipContent: 'Failed to process request' }}
    />
    <StatusWithDescription
      statusProps={{ status: 'warn', variant: 'timeline' }}
      tooltipProps={{
        tipContent: 'Warning: Check configuration',
        position: 'top',
      }}
    />
  </div>
)

export const AllVariants = () => (
  <div className="flex flex-col gap-4">
    <div className="flex items-center gap-4">
      <StatusWithDescription
        statusProps={{ status: 'default' }}
        tooltipProps={{ tipContent: 'Default status' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'success' }}
        tooltipProps={{ tipContent: 'Success status' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'error' }}
        tooltipProps={{ tipContent: 'Error status' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'warn' }}
        tooltipProps={{ tipContent: 'Warning status' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'info' }}
        tooltipProps={{ tipContent: 'Info status' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'brand' }}
        tooltipProps={{ tipContent: 'Brand status' }}
      />
    </div>
    <div className="flex items-center gap-4">
      <StatusWithDescription
        statusProps={{ status: 'default', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Default badge' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'active', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Active badge' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'error', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Error badge' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'warn', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Warning badge' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'info', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Info badge' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'brand', variant: 'badge' }}
        tooltipProps={{ tipContent: 'Brand badge' }}
      />
    </div>
    <div className="flex items-center gap-4">
      <StatusWithDescription
        statusProps={{ status: 'default', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Default timeline' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'success', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Success timeline' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'error', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Error timeline' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'warn', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Warning timeline' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'info', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Info timeline' }}
      />
      <StatusWithDescription
        statusProps={{ status: 'special', variant: 'timeline' }}
        tooltipProps={{ tipContent: 'Special timeline' }}
      />
    </div>
  </div>
)

export const TooltipPositions = () => (
  <div className="flex items-center justify-center gap-8 p-8">
    <StatusWithDescription
      statusProps={{ status: 'success', variant: 'badge' }}
      tooltipProps={{ tipContent: 'Top position', position: 'top' }}
    />
    <StatusWithDescription
      statusProps={{ status: 'error', variant: 'badge' }}
      tooltipProps={{ tipContent: 'Bottom position', position: 'bottom' }}
    />
    <StatusWithDescription
      statusProps={{ status: 'warn', variant: 'badge' }}
      tooltipProps={{ tipContent: 'Left position', position: 'left' }}
    />
    <StatusWithDescription
      statusProps={{ status: 'info', variant: 'badge' }}
      tooltipProps={{ tipContent: 'Right position', position: 'right' }}
    />
  </div>
)

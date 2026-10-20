export const isTerminalWorkflow = ({
  finished,
  status,
}: {
  finished?: boolean
  status?: string
}) =>
  !!finished ||
  ['success', 'error', 'cancelled', 'canceled'].includes(
    status?.toLowerCase() ?? ''
  )

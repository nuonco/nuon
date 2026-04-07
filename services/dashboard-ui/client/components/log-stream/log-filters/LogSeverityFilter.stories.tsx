export default {
  title: 'LogStream/LogSeverityFilter',
}

import { LogSeverityFilter } from './LogSeverityFilter'

export const Default = () => (
  <LogSeverityFilter
    title="severity"
    filters={{
      selectedSeverities: new Set(['Info', 'Warn', 'Error']),
      handleSeverityInputToggle: () => {},
      handleSeverityButtonClick: () => {},
      handleSeverityReset: () => {},
    }}
  />
)

export const AllSelected = () => (
  <LogSeverityFilter
    title="severity"
    filters={{
      selectedSeverities: new Set(['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']),
      handleSeverityInputToggle: () => {},
      handleSeverityButtonClick: () => {},
      handleSeverityReset: () => {},
    }}
  />
)

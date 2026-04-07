export default {
  title: 'LogStream/LogServiceFilter',
}

import { LogServiceFilter } from './LogServiceFilter'

export const Default = () => (
  <LogServiceFilter
    title="service"
    filters={{
      selectedServices: new Set(['api', 'runner']),
      handleServiceInputToggle: () => {},
      handleServiceButtonClick: () => {},
      handleServiceReset: () => {},
    }}
  />
)

export const NoneSelected = () => (
  <LogServiceFilter
    title="service"
    filters={{
      selectedServices: new Set(),
      handleServiceInputToggle: () => {},
      handleServiceButtonClick: () => {},
      handleServiceReset: () => {},
    }}
  />
)

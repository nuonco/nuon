export default {
  title: 'LogStream/LogJobOutputFilter',
}

import { useState } from 'react'
import { LogJobOutputFilter } from './LogJobOutputFilter'

export const Default = () => {
  const [jobOutputOnly, setJobOutputOnly] = useState(false)
  return (
    <LogJobOutputFilter
      filters={{
        jobOutputOnly,
        handleJobOutputToggle: () => setJobOutputOnly((v) => !v),
      }}
    />
  )
}

export const Checked = () => (
  <LogJobOutputFilter
    filters={{
      jobOutputOnly: true,
      handleJobOutputToggle: () => {},
    }}
  />
)

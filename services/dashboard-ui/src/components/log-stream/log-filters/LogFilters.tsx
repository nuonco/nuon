import { LogSearch } from './LogSearch'
import { LogServiceFilter } from './LogServiceFilter'
import { LogSeverityFilter } from './LogSeverityFilter'
import { LogSort } from './LogSort'

export const LogFilters = () => {
  return (
    <div className="flex flex-wrap items-center justify-between gap-4 py-4">
      <LogSearch />

      <div className="flex items-center gap-4">
        <LogSort />
        <LogServiceFilter title="service" />
        <LogSeverityFilter title="severity" />
      </div>
    </div>
  )
}

export default {
  title: 'LogStream/LogLineSkeleton',
}

import { LogLineSkeleton } from './LogLine'

export const Default = () => (
  <div className="p-4">
    <LogLineSkeleton />
  </div>
)

export const Multiple = () => (
  <div className="p-4 flex flex-col">
    {Array.from({ length: 5 }).map((_, idx) => (
      <LogLineSkeleton key={idx} />
    ))}
  </div>
)

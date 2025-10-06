import { TimelineSkeleton } from './TimelineSkeleton'

export const Default = () => <TimelineSkeleton />

export const WithCustomEventCount = () => (
  <div className="flex flex-col gap-8">
    <div>
      <h3 className="mb-4 text-lg font-semibold">3 Events</h3>
      <TimelineSkeleton eventCount={3} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">8 Events</h3>
      <TimelineSkeleton eventCount={8} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">1 Event</h3>
      <TimelineSkeleton eventCount={1} />
    </div>
  </div>
)

export const WithCustomClassName = () => (
  <TimelineSkeleton className="border border-gray-200 rounded-lg p-4" />
)

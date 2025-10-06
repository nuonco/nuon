import { Duration } from './Duration'

export const Default = () => (
  <div className="flex flex-col gap-4">
    <Duration nanoseconds={1000000000} />
    <Duration beginTime="2022-01-01T00:00:00Z" endTime="2022-01-01T00:00:01Z" />
  </div>
)

export const Formats = () => (
  <div className="flex flex-col gap-4">
    <Duration nanoseconds={1000000000} format="default" />
    <Duration nanoseconds={1000000000} format="timer" />
  </div>
)

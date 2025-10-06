import { Time } from './Time'

export const Default = () => <Time />

export const Formats = () => (
  <div className="flex flex-col gap-4">
    <Time format="short-datetime" />
    <Time format="long-datetime" />
    <Time format="relative" />
    <Time format="time-only" />
    <Time format="log-datetime" />
  </div>
)

export const WithSeconds = () => (
  <div className="flex flex-col gap-4">
    <div>
      <strong>Unix timestamp (1640995200 = Jan 1, 2022):</strong>
      <Time seconds={1640995200} />
    </div>
    <div>
      <strong>Recent timestamp (1 hour ago):</strong>
      <Time seconds={Math.floor(Date.now() / 1000) - 3600} />
    </div>
    <div>
      <strong>Future timestamp (in 2 hours):</strong>
      <Time seconds={Math.floor(Date.now() / 1000) + 7200} />
    </div>
  </div>
)

export const SecondsWithFormats = () => (
  <div className="flex flex-col gap-4">
    <div>
      <strong>Short datetime format:</strong>
      <Time seconds={1640995200} format="short-datetime" />
    </div>
    <div>
      <strong>Long datetime format:</strong>
      <Time seconds={1640995200} format="long-datetime" />
    </div>
    <div>
      <strong>Relative format:</strong>
      <Time seconds={1640995200} format="relative" />
    </div>
    <div>
      <strong>Time only:</strong>
      <Time seconds={1640995200} format="time-only" />
    </div>
    <div>
      <strong>Log datetime format:</strong>
      <Time seconds={1640995200} format="log-datetime" />
    </div>
  </div>
)

export const ComparisonTimestamp = () => {
  const timestamp = 1640995200
  const isoString = '2022-01-01T00:00:00.000Z'

  return (
    <div className="flex flex-col gap-4">
      <div>
        <strong>From seconds ({timestamp}):</strong>
        <Time seconds={timestamp} />
      </div>
      <div>
        <strong>From ISO string ({isoString}):</strong>
        <Time time={isoString} />
      </div>
    </div>
  )
}

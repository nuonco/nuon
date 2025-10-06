import { Status } from './Status'

export const Variants = () => (
  <div className="flex flex-col gap-4">
    <Status status="default" />
    <Status status="success" variant="badge" />
    <Status status="error" variant="timeline" />
  </div>
)

export const All = () => (
  <div className="flex flex-col gap-4">
    <div className="flex items-center gap-4">
      <Status status="default" />
      <Status status="success" />
      <Status status="error" />
      <Status status="warn" />
      <Status status="info" />
      <Status status="brand" />
    </div>
    <div className="flex items-center gap-4">
      <Status status="default" variant="badge" />
      <Status status="active" variant="badge" />
      <Status status="error" variant="badge" />
      <Status status="warn" variant="badge" />
      <Status status="info" variant="badge" />
      <Status status="brand" variant="badge" />
    </div>
    <div className="flex items-center gap-4">
      <Status status="default" variant="timeline" />
      <Status status="success" variant="timeline" />
      <Status status="error" variant="timeline" />
      <Status status="warn" variant="timeline" />
      <Status status="info" variant="timeline" />
      <Status status="special" variant="timeline" />
    </div>
  </div>
)

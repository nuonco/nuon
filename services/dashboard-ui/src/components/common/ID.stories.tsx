import { ID } from './ID'

export const Default = () => (
  <div className="flex flex-col gap-4">
    <ID>abc-123-def</ID>
    <ID>user-456789</ID>
    <ID>app_config_12345</ID>
  </div>
)

export const DifferentIDTypes = () => (
  <div className="flex flex-col gap-4">
    <ID>UUID: f47ac10b-58cc-4372-a567-0e02b2c3d479</ID>
    <ID>Short ID: abc123</ID>
    <ID>Database ID: 1234567890</ID>
    <ID>Hash: sha256:a3b5c2d7e9f1234567890abcdef</ID>
  </div>
)

export const WithTextProps = () => (
  <div className="flex flex-col gap-4">
    <ID variant="subtext">Subtext ID</ID>
    <ID variant="base">Base ID</ID>
    <ID variant="h3">H3 ID</ID>
    <ID variant="h1">H1 ID</ID>
  </div>
)

export const WithClickToCopyProps = () => (
  <div className="flex flex-col gap-4">
    <ID clickToCopyProps={{ className: 'bg-blue-50 p-2 rounded' }}>
      styled-container-id
    </ID>
    <ID clickToCopyProps={{ noticeClassName: '!bg-green-500 text-white' }}>
      custom-notice-id
    </ID>
  </div>
)

export const LongIDs = () => (
  <div className="flex flex-col gap-4 max-w-md">
    <ID>very-long-identifier-that-might-wrap-in-containers</ID>
    <ID>arn:aws:iam::123456789012:role/service-role/MyLambdaRole</ID>
    <ID>projects/my-project/locations/us-central1/instances/my-instance</ID>
  </div>
)

import { ComponentDocs } from '../__stories__/ComponentDocs'
import { FormErrorBanner } from './FormErrorBanner'

export default { title: 'lite/molecules/FormErrorBanner' }

export const Overview = () => (
  <ComponentDocs
    name="FormErrorBanner"
    tier="molecule"
    summary="An error Banner that reads a submission failure straight off an API error."
    use={[
      'Place inside the form above its actions.',
      'Show server errors that are not owned by one field.',
    ]}
    avoid={[
      'Do not use a toast for form submission errors. The toast leaves while the broken form stays.',
      'Do not use for field validation. A field owns its own error.',
      'Do not reach for a plain Banner in a form. This one already knows the API error shape.',
    ]}
    rules={[
      'The API error heading wins over the thrown message, which wins over fallback.',
      'Description adds server context when present.',
      'It renders an error Banner, so it announces assertively and needs no role of its own.',
      'A null or undefined error renders nothing, so a caller can pass a query error straight through.',
    ]}
    props={[
      {
        name: 'error',
        type: 'TAPIError | Error | null | undefined',
        description: 'Error to display; null or undefined renders nothing.',
      },
      {
        name: 'fallback',
        type: 'string',
        description: 'Message used when the error has no useful text.',
      },
    ]}
  />
)

export const ApiError = () => (
  <div className="max-w-lg p-8">
    <FormErrorBanner
      fallback="Unable to save changes"
      error={{
        error: 'Configuration update failed',
        description: 'The configuration changed after this form was opened.',
        user_error: true,
      }}
    />
  </div>
)

export const JavaScriptError = () => (
  <div className="max-w-lg p-8">
    <FormErrorBanner
      fallback="Unable to save changes"
      error={new Error('Network connection lost')}
    />
  </div>
)

export const Fallback = () => (
  <div className="max-w-lg p-8">
    <FormErrorBanner fallback="Unable to save changes" error={new Error()} />
  </div>
)

export const NoError = () => (
  <div className="max-w-lg p-8">
    <FormErrorBanner fallback="Unable to save changes" error={null} />
  </div>
)

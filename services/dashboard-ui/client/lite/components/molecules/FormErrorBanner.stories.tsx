import { ComponentDocs } from '../__stories__/ComponentDocs'
import { FormErrorBanner } from './FormErrorBanner'

export default { title: 'lite/molecules/FormErrorBanner' }

export const Overview = () => (
  <ComponentDocs
    name="FormErrorBanner"
    tier="molecule"
    summary="A form-level API error surface built from Card and Text."
    use={[
      'Place inside the form above its actions.',
      'Show server errors that are not owned by one field.',
    ]}
    avoid={[
      'Do not use a toast for form submission errors.',
      'Do not use for field validation.',
    ]}
    rules={[
      'The API error heading wins over fallback.',
      'Description adds server context when present.',
    ]}
    props={[
      {
        name: 'error',
        type: 'TAPIError | Error | null',
        description: 'Error to display; null renders nothing.',
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

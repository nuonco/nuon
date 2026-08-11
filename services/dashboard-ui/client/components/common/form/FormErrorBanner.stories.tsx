import type { TAPIError } from '@/types'
import { FormErrorBanner } from './FormErrorBanner'

export default { title: 'Common/Forms/FormErrorBanner' }

const apiError = { error: 'A webhook with this URL already exists.' } as TAPIError

export const WithError = () => (
  <div className="max-w-md p-4">
    <FormErrorBanner error={apiError} fallback="Unable to create webhook" />
  </div>
)

const apiErrorWithDescription = {
  error: 'this action requires write access to installs in this organization',
  description:
    'Your role (App developer) does not have write access to installs in this organization. Ask an organization admin to assign a role that does.',
} as TAPIError

export const WithDescription = () => (
  <div className="max-w-md p-4">
    <FormErrorBanner
      error={apiErrorWithDescription}
      fallback="Unable to perform action"
    />
  </div>
)

export const FallbackMessage = () => (
  <div className="max-w-md p-4">
    <FormErrorBanner error={{} as TAPIError} fallback="Unable to create webhook" />
  </div>
)

export const NoError = () => (
  <div className="max-w-md p-4">
    <FormErrorBanner error={null} fallback="Unable to create webhook" />
  </div>
)

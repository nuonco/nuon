export default {
  title: 'Layout/ProviderError',
}

import { ProviderError } from './ProviderError'
import type { TAPIError } from '@/types'

const notFoundError: TAPIError = {
  status: 404,
  error: 'The resource you requested could not be found.',
  description: 'Not found',
  user_error: false,
}

const serverError: TAPIError = {
  status: 500,
  error: 'An internal server error occurred. Please try again later.',
  description: 'Internal server error',
  user_error: false,
}

export const NotFound = () => <ProviderError error={notFoundError} />

export const ServerError = () => <ProviderError error={serverError} />

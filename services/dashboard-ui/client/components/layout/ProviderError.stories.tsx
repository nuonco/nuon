export default {
  title: 'Layout/ProviderError',
}

import { ProviderError } from './ProviderError'
import type { TAPIError } from '@/types'

const notFoundError: TAPIError = {
  status: 404,
  error: 'The resource you requested could not be found.',
}

const serverError: TAPIError = {
  status: 500,
  error: 'An internal server error occurred. Please try again later.',
}

export const NotFound = () => <ProviderError error={notFoundError} />

export const ServerError = () => <ProviderError error={serverError} />

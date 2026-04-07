export default {
  title: 'Layout/AuthLayout',
}

import { AuthLayout } from './AuthLayout'

export const Loading = () => (
  <AuthLayout isLoading isAuthenticated={false} hasError={false} onRetry={() => {}} />
)

export const Unauthenticated = () => (
  <AuthLayout isLoading={false} isAuthenticated={false} hasError={false} onRetry={() => {}} />
)

export const Error = () => (
  <AuthLayout isLoading={false} isAuthenticated={false} hasError onRetry={() => {}} />
)

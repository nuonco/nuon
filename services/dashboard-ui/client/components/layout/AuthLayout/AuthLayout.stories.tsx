export default {
  title: 'UI / Layout / Auth layout',
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

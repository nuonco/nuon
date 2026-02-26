import { BrowserRouter, Routes, Route } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ConfigProvider } from '@/providers/config-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Login } from '@/views/Login'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

const Root = () => <>root</>
const Apps = () => <>Apps</>

export const App = () => {
  return (
    <ConfigProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route element={<AuthLayout />}>
                <Route path="/" element={<Root />} />
                <Route path="/apps" element={<Apps />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </AuthProvider>
        <ReactQueryDevtools />
      </QueryClientProvider>
    </ConfigProvider>
  )
}

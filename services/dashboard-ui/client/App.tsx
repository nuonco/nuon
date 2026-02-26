import { BrowserRouter, Routes, Route } from 'react-router'
import { ConfigProvider } from '@/providers/config-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Login } from '@/views/Login'

const Root = () => <>root</>
const Apps = () => <>Apps</>

export const App = () => {
  return (
    <ConfigProvider>
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
    </ConfigProvider>
  )
}

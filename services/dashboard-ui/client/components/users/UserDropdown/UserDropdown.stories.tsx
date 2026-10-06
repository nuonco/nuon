export default {
  title: 'Features / Users / User dropdown',
}

import { UserDropdown } from './UserDropdown'
import { ThemeProvider } from '@/providers/theme-provider'

export const Default = () => (
  <ThemeProvider>
    <div className="p-4">
      <UserDropdown
        isByoc={false}
        isAdmin={false}
        isNuonEmployee={false}
        isDev={false}
        apiUrl="https://api.nuon.co"
        authServiceUrl="https://auth.nuon.co"
        notificationsSupported={false}
        notificationPermission="default"
        muted={false}
        onToggleMute={() => {}}
        onRequestPermission={async () => 'denied'}
        onAddPanel={() => ''}
        onAddToast={() => ''}
        user={{ name: 'Jane Smith', email: 'jane@example.com' }}
        isUserLoading={false}
      />
    </div>
  </ThemeProvider>
)

export const Admin = () => (
  <ThemeProvider>
    <div className="p-4">
      <UserDropdown
        isByoc={false}
        isAdmin
        isNuonEmployee
        isDev={false}
        apiUrl="https://api.nuon.co"
        adminDashboardUrl="https://admin.nuon.co"
        grafanaUiUrl="https://app.nuon.co/admin/grafana"
        authServiceUrl="https://auth.nuon.co"
        notificationsSupported
        notificationPermission="granted"
        muted={false}
        onToggleMute={() => {}}
        onRequestPermission={async () => 'granted'}
        onAddPanel={() => ''}
        onAddToast={() => ''}
        user={{ name: 'Admin User', email: 'admin@nuon.co' }}
        isUserLoading={false}
      />
    </div>
  </ThemeProvider>
)

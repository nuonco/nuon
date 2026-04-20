export default {
  title: 'Users/UserDropdown',
}

import { UserDropdown } from './UserDropdown'

export const Default = () => (
  <div className="p-4">
    <UserDropdown
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
)

export const Admin = () => (
  <div className="p-4">
    <UserDropdown
      isAdmin
      isNuonEmployee
      isDev={false}
      apiUrl="https://api.nuon.co"
      adminDashboardUrl="https://admin.nuon.co"
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
)

import { useParams } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { AdminControls } from './AdminControls'

export const AdminControlsContainer = () => {
  const { isNuonEmployee } = useAuth()
  const params = useParams()
  const orgId = params?.orgId as string
  const appId = params?.appId as string | undefined
  const installId = params?.installId as string | undefined

  return (
    <AdminControls
      isNuonEmployee={isNuonEmployee}
      orgId={orgId}
      appId={appId}
      installId={installId}
    />
  )
}

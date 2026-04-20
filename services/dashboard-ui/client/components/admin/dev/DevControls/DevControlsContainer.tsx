import { useParams } from 'react-router'
import { DevControls } from './DevControls'

export const DevControlsContainer = () => {
  const params = useParams()
  const orgId = params?.orgId as string
  const installId = params?.installId as string | undefined

  return <DevControls orgId={orgId} installId={installId} />
}

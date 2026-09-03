import { Navigate, useParams } from 'react-router'
import { appById } from './fixtures'

export const AppResolver = () => {
  const { appId = '' } = useParams()
  const branch = appById(appId).branches.at(0)

  return (
    <Navigate
      replace
      to={
        branch ? `/apps/${appId}/branches/${branch.id}` : `/apps/${appId}/setup`
      }
    />
  )
}

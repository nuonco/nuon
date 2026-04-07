import { useSearchParams } from 'react-router'
import { useEffect } from 'react'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { VCSConnectionSuccessModal } from './VCSConnectionSuccess'

export const VCSConnectionSuccess = ({}: {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const vcsId = searchParams?.get('vcs-connected')
  const vcs_connection = org?.vcs_connections?.find((vcs) => vcs?.id === vcsId)

  useEffect(() => {
    if (vcsId) {
      const modal = (
        <VCSConnectionSuccessModal
          orgName={org?.name}
          vcs_connection={vcs_connection}
        />
      )
      addModal(modal)
    }
  }, [])

  return <></>
}

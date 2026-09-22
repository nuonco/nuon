import { useOrg } from '@/hooks/use-org'

export const useNewInstallIA = () => {
  const { org } = useOrg()
  return (
    !!org?.features?.['app-branches-ui'] && !!org?.features?.['new-install-ia']
  )
}

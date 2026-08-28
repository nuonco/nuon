import { useOrg } from '@/hooks/use-org'

export const useNewAppIA = () => {
  const { org } = useOrg()
  return !!org?.features?.['app-branches-ui'] && !!org?.features?.['new-app-ia']
}

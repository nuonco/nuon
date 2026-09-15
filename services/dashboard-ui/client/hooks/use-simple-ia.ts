import { useOrg } from '@/hooks/use-org'

export const useSimpleIA = () => {
  const { org } = useOrg()
  return !!org?.features?.['simple-ia']
}

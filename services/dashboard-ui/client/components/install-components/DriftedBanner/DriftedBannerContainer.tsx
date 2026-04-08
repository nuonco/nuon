import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { DriftedBanner } from './DriftedBanner'
import type { TDriftedObject } from '@/types'

export const DriftedBannerContainer = ({ drifted }: { drifted: TDriftedObject }) => {
  const { org } = useOrg()
  const { install } = useInstall()

  return <DriftedBanner drifted={drifted} orgId={org.id} installId={install.id} />
}

export default {
  title: 'Features / Builds / Signature badge',
}

import type { TBuild } from '@/types'
import { BuildSignatureBadge } from './BuildSignatureBadge'

const build = (b: Partial<TBuild>) => b as unknown as TBuild

export const States = () => (
  <div className="flex flex-wrap gap-2">
    <BuildSignatureBadge
      build={build({ signature_verification: 'verified' })}
    />
    <BuildSignatureBadge
      build={build({ signature_verification: 'rejected' })}
    />
    <BuildSignatureBadge
      build={build({ signature_verification: 'not_required' })}
    />
    <BuildSignatureBadge build={build({ status: 'error' })} />
  </div>
)

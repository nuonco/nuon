export default {
  title: 'Builds/BuildImageSource',
}

import type { TBuild } from '@/types'
import { BuildImageSource } from './BuildImageSource'

const build = {
  source_ref: 'nginx:1.25.5',
  resolved_tag: '1.25.5',
  source_digest:
    'sha256:9f1c4a0b1e2d3c4b5a6978d0e1f2a3b4c5d6e7f8091a2b3c4d5e6f7081920a3b',
  source_media_type: 'application/vnd.oci.image.index.v1+json',
  resolved_at: '2026-08-24T10:12:00Z',
} as unknown as TBuild

export const Default = () => <BuildImageSource build={build} />

export const Partial = () => (
  <BuildImageSource
    build={{ source_ref: 'nginx:1.25.5' } as unknown as TBuild}
  />
)

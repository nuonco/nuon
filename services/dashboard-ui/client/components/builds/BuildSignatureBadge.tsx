import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Tooltip } from '@/components/common/Tooltip'
import type { TBuild } from '@/types'

const SIGNATURE_STATES: Record<
  string,
  { label: string; theme: TBadgeTheme; tip: string }
> = {
  verified: {
    label: 'Signature verified',
    theme: 'success',
    tip: 'The image signature matched a trusted authority before the image was copied.',
  },
  rejected: {
    label: 'Signature rejected',
    theme: 'error',
    tip: 'No trusted authority verified the image signature, so the build failed.',
  },
  not_required: {
    label: 'Unverified',
    theme: 'neutral',
    tip: 'Signature verification is not required for this component.',
  },
  unknown: {
    label: 'Signature unknown',
    theme: 'neutral',
    tip: 'The build failed before the image signature was checked.',
  },
}

const signatureState = (build: TBuild) => {
  if (build?.signature_verification) return build.signature_verification
  return build?.status_v2?.status === 'error' || build?.status === 'error'
    ? 'unknown'
    : undefined
}

export const BuildSignatureBadge = ({ build }: { build: TBuild }) => {
  const state = SIGNATURE_STATES[signatureState(build) ?? '']
  if (!state) return null

  return (
    <Tooltip
      tipContent={state.tip}
      tipContentClassName="!whitespace-normal !w-auto max-w-[240px] text-xs"
    >
      <Badge size="sm" theme={state.theme}>
        {state.label}
      </Badge>
    </Tooltip>
  )
}

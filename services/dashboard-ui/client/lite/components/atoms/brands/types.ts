import type { SVGAttributes } from 'react'

export type TBrandTone = 'color' | 'mono'

export interface IBrandMark
  extends Omit<SVGAttributes<SVGSVGElement>, 'color'> {
  size?: number | string
  tone?: TBrandTone
}

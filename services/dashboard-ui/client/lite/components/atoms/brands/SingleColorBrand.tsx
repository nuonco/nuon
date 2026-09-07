import type { IBrandMark } from './types'

interface ISingleColorBrand extends IBrandMark {
  color: string
  path: string
  viewBox?: string
}

export const SingleColorBrand = ({
  color,
  path,
  size = 16,
  tone = 'color',
  viewBox = '0 0 24 24',
  style,
  ...props
}: ISingleColorBrand) => (
  <svg
    viewBox={viewBox}
    width={size}
    height={size}
    fill="currentColor"
    style={{ color: tone === 'color' ? color : undefined, ...style }}
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <path d={path} />
  </svg>
)

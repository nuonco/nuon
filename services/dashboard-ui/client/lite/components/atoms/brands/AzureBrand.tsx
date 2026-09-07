import { useId } from 'react'
import type { IBrandMark } from './types'

export const AzureBrand = ({
  size = 16,
  tone = 'color',
  ...props
}: IBrandMark) => {
  const id = useId()

  if (tone === 'mono') {
    return (
      <svg
        viewBox="0 0 96 96"
        width={size}
        height={size}
        fill="currentColor"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path d="M33.338 6.544h26.038l-27.03 80.455a4.46 4.46 0 0 1-4.24 3.046H12.378a4.474 4.474 0 0 1-4.225-5.963L31.1 9.59a4.46 4.46 0 0 1 4.239-3.046zM71.175 60.261H41.64a1.912 1.912 0 0 0-1.305 3.309l26.532 24.764a4.567 4.567 0 0 0 3.122 1.229h18.476zM64.933 9.59a4.46 4.46 0 0 0-4.24-3.046H33.79a4.46 4.46 0 0 1 4.239 3.046L60.976 84.082a4.474 4.474 0 0 1-4.225 5.963h26.903a4.474 4.474 0 0 0 4.225-5.963z" />
      </svg>
    )
  }

  const gradientA = `${id}-azure-a`
  const gradientB = `${id}-azure-b`
  const gradientC = `${id}-azure-c`

  return (
    <svg
      viewBox="0 0 96 96"
      width={size}
      height={size}
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <defs>
        <linearGradient
          id={gradientA}
          x1={58.972}
          x2={37.191}
          y1={9.822}
          y2={100.666}
          gradientUnits="userSpaceOnUse"
        >
          <stop offset={0} stopColor="#114A8B" />
          <stop offset={1} stopColor="#0669BC" />
        </linearGradient>
        <linearGradient
          id={gradientB}
          x1={60.089}
          x2={53.512}
          y1={52.014}
          y2={54.48}
          gradientUnits="userSpaceOnUse"
        >
          <stop offset={0} stopOpacity={0.3} />
          <stop offset={0.071} stopOpacity={0.2} />
          <stop offset={0.321} stopOpacity={0.1} />
          <stop offset={0.623} stopOpacity={0.05} />
          <stop offset={1} stopOpacity={0} />
        </linearGradient>
        <linearGradient
          id={gradientC}
          x1={46.112}
          x2={64.81}
          y1={4.063}
          y2={98.645}
          gradientUnits="userSpaceOnUse"
        >
          <stop offset={0} stopColor="#3CCBF4" />
          <stop offset={1} stopColor="#2892DF" />
        </linearGradient>
      </defs>
      <path
        d="M33.338 6.544h26.038l-27.03 80.455a4.46 4.46 0 0 1-4.24 3.046H12.378a4.474 4.474 0 0 1-4.225-5.963L31.1 9.59a4.46 4.46 0 0 1 4.239-3.046z"
        fill={`url(#${gradientA})`}
      />
      <path
        d="M71.175 60.261H41.64a1.912 1.912 0 0 0-1.305 3.309l26.532 24.764a4.567 4.567 0 0 0 3.122 1.229h18.476z"
        fill="#0078D4"
      />
      <path
        d="M33.338 6.544a4.42 4.42 0 0 0-4.226 3.12L6.195 84.074a4.462 4.462 0 0 0 4.188 6.007h16.399a4.543 4.543 0 0 0 3.626-2.585l5.407-13.942 19.472 14.958a4.536 4.536 0 0 0 2.604 1.072h18.629l-8.134-24.322-25.26.003L59.56 6.544z"
        fill={`url(#${gradientB})`}
      />
      <path
        d="M64.933 9.59a4.46 4.46 0 0 0-4.24-3.046H33.79a4.46 4.46 0 0 1 4.239 3.046L60.976 84.082a4.474 4.474 0 0 1-4.225 5.963h26.903a4.474 4.474 0 0 0 4.225-5.963z"
        fill={`url(#${gradientC})`}
      />
    </svg>
  )
}

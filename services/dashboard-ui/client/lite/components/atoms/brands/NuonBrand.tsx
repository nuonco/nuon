import { useId } from 'react'
import type { IBrandMark } from './types'

export const NuonBrand = ({
  size = 16,
  tone = 'color',
  ...props
}: IBrandMark) => {
  const gradientId = `${useId()}-nuon`

  return (
    <svg
      viewBox="0 0 20.229 28"
      width={size}
      height={size}
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <defs>
        <linearGradient
          id={gradientId}
          x1="-1.26608e-07"
          y1="25.0883"
          x2="10.9568"
          y2="-0.844511"
          gradientUnits="userSpaceOnUse"
        >
          <stop stopColor="#F72585" />
          <stop offset="0.53" stopColor="#3A00FF" />
          <stop offset="1" stopColor="#4CC9F0" />
        </linearGradient>
      </defs>
      <path
        d="M14.8701 0L9.5113 3.09463V8.10487L5.17338 5.59852H5.1709L0 8.58439V25.0141L5.16843 28H5.1709L10.72 24.7941V20.0039L14.8701 22.399L20.2288 19.3044V3.09463L14.8701 0ZM1.21116 9.2839L5.16843 7H5.1709L9.50883 9.50388V17.9054L1.21116 13.1151V9.2839ZM9.50883 24.0946L5.16843 26.5985L1.21116 24.3146V14.5141L9.50883 19.3044V24.0946ZM19.0177 18.6024L14.8701 20.9975L10.7225 18.6049V10.2034L19.0201 14.9936V18.6024H19.0177ZM19.0177 13.5946L10.72 8.80438V3.79414L14.8701 1.39901L19.0177 3.79414V13.5946Z"
        fill={tone === 'color' ? `url(#${gradientId})` : 'currentColor'}
      />
    </svg>
  )
}

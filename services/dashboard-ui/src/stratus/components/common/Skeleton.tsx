import React from 'react'

interface ISkeleton {
  lines?: number
  width?: string | string[]
  height?: string
}

export const Skeleton = ({
  lines = 1,
  width = '100%',
  height = '1rem',
}: ISkeleton) => {
  const widths = Array.isArray(width) ? width : Array(lines).fill(width)

  return (
    <div className="skeleton-text">
      {Array.from({ length: lines }).map((_, index) => (
        <div
          key={index}
          className="animate-pulse rounded-sm bg-cool-grey-400 dark:bg-dark-grey-400"
          style={{
            width: widths[index] || '100%',
            height: height,
          }}
        ></div>
      ))}
    </div>
  )
}

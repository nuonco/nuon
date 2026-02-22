import React from 'react'

interface NextImageProps {
  src: string
  alt?: string
  width?: number
  height?: number
  className?: string
  priority?: boolean
  placeholder?: string
  blurDataURL?: string
  fill?: boolean
  sizes?: string
  quality?: number
  [key: string]: any
}

const Image = React.forwardRef<HTMLImageElement, NextImageProps>(
  (
    { src, alt, width, height, className, priority, placeholder, blurDataURL, fill, sizes, quality, ...props },
    ref
  ) => {
    const style: React.CSSProperties = fill
      ? { position: 'absolute', inset: 0, width: '100%', height: '100%', objectFit: 'cover' }
      : {}

    return React.createElement('img', {
      ref,
      src,
      alt: alt || '',
      width: fill ? undefined : width,
      height: fill ? undefined : height,
      className,
      style: fill ? style : undefined,
      loading: priority ? 'eager' : 'lazy',
      referrerPolicy: 'no-referrer',
      ...props,
    })
  }
)

Image.displayName = 'Image'

export default Image

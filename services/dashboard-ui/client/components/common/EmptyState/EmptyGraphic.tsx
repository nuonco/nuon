import { useEffect, useId, useMemo, useState } from 'react'
import type { TEmptyVariant } from '@/types'
import { cn } from '@/utils/classnames'

interface IEmptyGraphic {
  isDarkModeOnly?: boolean
  size?: 'default' | 'sm'
  variant?: TEmptyVariant
}

const resolved = new Map<string, string>()
const pending = new Map<string, Promise<string>>()

function loadSvg(src: string) {
  const hit = resolved.get(src)
  if (hit) return Promise.resolve(hit)

  let job = pending.get(src)
  if (!job) {
    job = fetch(src).then(async (response) => {
      if (!response.ok) throw new Error(`Failed to load ${src}`)
      const text = await response.text()
      resolved.set(src, text)
      return text
    })
    pending.set(src, job)
  }
  return job
}

function useSvg(src: string) {
  const [svg, setSvg] = useState<string | undefined>(() => resolved.get(src))

  useEffect(() => {
    const cached = resolved.get(src)
    if (cached) {
      setSvg(cached)
      return
    }

    let cancelled = false
    loadSvg(src).then(
      (text) => {
        if (!cancelled) setSvg(text)
      },
      () => {}
    )

    return () => {
      cancelled = true
    }
  }, [src])

  return svg
}

function scopeSvg(svg: string, prefix: string) {
  const ids = [...svg.matchAll(/\bid="([^"]+)"/g)].map((match) => match[1])
  ids.sort((a, b) => b.length - a.length)

  let next = svg
  for (const id of ids) {
    const scoped = `${prefix}-${id}`
    next = next
      .replaceAll(`id="${id}"`, `id="${scoped}"`)
      .replaceAll(`url(#${id})`, `url(#${scoped})`)
      .replaceAll(`href="#${id}"`, `href="#${scoped}"`)
  }

  return next.replace('<svg ', '<svg class="block h-full w-full" ')
}

const Graphic = ({
  className,
  prefix,
  src,
}: {
  className: string
  prefix: string
  src: string
}) => {
  const svg = useSvg(src)
  const html = useMemo(() => (svg ? scopeSvg(svg, prefix) : ''), [prefix, svg])

  return (
    <span
      aria-hidden
      className={className}
      {...(html ? { dangerouslySetInnerHTML: { __html: html } } : {})}
    />
  )
}

export const EmptyGraphic = ({
  isDarkModeOnly = false,
  size = 'default',
  variant = '404',
}: IEmptyGraphic) => {
  const sizeSuffix = size === 'sm' ? '-small' : ''
  const sizeClass = size === 'sm' ? 'h-[64px] w-[94px]' : 'h-[94px] w-[154px]'
  const idPrefix = `eg${useId().replace(/[^A-Za-z0-9]/g, '')}`

  return (
    <>
      <Graphic
        className={cn(sizeClass, 'relative block', {
          hidden: isDarkModeOnly,
          'dark:hidden': !isDarkModeOnly,
        })}
        prefix={`${idPrefix}l`}
        src={`/empty-graphics/${variant}-light${sizeSuffix}.svg`}
      />
      <Graphic
        className={cn(sizeClass, 'relative dark:block', {
          block: isDarkModeOnly,
          hidden: !isDarkModeOnly,
        })}
        prefix={`${idPrefix}d`}
        src={`/empty-graphics/${variant}-dark${sizeSuffix}.svg`}
      />
    </>
  )
}

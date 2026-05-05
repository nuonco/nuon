import React, { lazy, Suspense, useEffect, useRef } from 'react'
import { Skeleton } from '@/components/common/Skeleton'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

const JsonViewer = lazy(
  () => import('@andypf/json-viewer/dist/esm/react/JsonViewer')
) as typeof import('@andypf/json-viewer/dist/esm/react/JsonViewer')

export interface IJSONViewer extends React.HTMLAttributes<HTMLDivElement> {
  data: any
  expanded?: number | boolean
  indent?: number
  showDataTypes?: boolean
  showToolbar?: boolean
  showCopy?: boolean
  showSize?: boolean
  expandIconType?: 'square' | 'circle' | 'arrow'
}

export const JSONViewer = ({
  className,
  data,
  expanded = 2,
  indent = 2,
  showDataTypes = true,
  showToolbar = false,
  showCopy = true,
  showSize = true,
  expandIconType = 'square',
  ...props
}: IJSONViewer) => {
  const colorScheme = useSystemTheme()
  const containerRef = useRef<HTMLDivElement>(null)

  const customDarkTheme = {
    base00: '#19171C',
    base01: '#19171C',
    base02: '#3E4451',
    base03: '#5C6370',
    base04: '#9DA0A2',
    base05: '#ABB2BF',
    base06: '#ABB2BF',
    base07: '#E06C75',
    base08: '#E06C75',
    base09: '#98C379',
    base0A: '#E5C07B',
    base0B: '#61AFEF',
    base0C: '#D19A66',
    base0D: '#56B6C2',
    base0E: '#C678DD',
    base0F: '#D19A66',
  }

  const customLightTheme = {
    base00: '#F0F3F5',
    base01: '#F0F3F5',
    base02: '#E5E5E6',
    base03: '#A0A1A7',
    base04: '#696C77',
    base05: '#383A42',
    base06: '#383A42',
    base07: '#E45649',
    base08: '#E45649',
    base09: '#50A14F',
    base0A: '#C18401',
    base0B: '#4078F2',
    base0C: '#C18401',
    base0D: '#0184BC',
    base0E: '#A626A4',
    base0F: '#C18401',
  }

  const theme = colorScheme === 'dark' ? customDarkTheme : customLightTheme
  const themeString = JSON.stringify(theme)

  useEffect(() => {
    const injectShadowCSS = () => {
      if (!containerRef.current) return

      const jsonViewer =
        containerRef.current.querySelector('andypf-json-viewer')
      if (!jsonViewer || !jsonViewer.shadowRoot) return

      const sheet = new CSSStyleSheet()
      const css = `
        .container,
        .container * {
          font-family: var(--font-hack) !important;
          font-size: 0.875rem !important;
          line-height: 1.25rem !important;
        }
        .key-value-wrapper,
        .value-string,
        .key-clickable,
        .colon,
        .data-row {
          font-family: var(--font-hack) !important;
          font-size: 0.875rem !important;
        }
      `

      sheet.replaceSync(css)

      if (jsonViewer.shadowRoot.adoptedStyleSheets) {
        jsonViewer.shadowRoot.adoptedStyleSheets = [
          ...jsonViewer.shadowRoot.adoptedStyleSheets,
          sheet,
        ]
      }
    }

    const timer = setTimeout(injectShadowCSS, 100)

    return () => clearTimeout(timer)
  }, [data, colorScheme])

  // The underlying web component copies string values via JSON.stringify, which
  // wraps them in quotes. Intercept the click on per-row copy icons inside the
  // shadow DOM and write the raw string (without the surrounding quotes).
  useEffect(() => {
    if (!containerRef.current) return

    let cleanup: (() => void) | undefined

    const attach = () => {
      const jsonViewer =
        containerRef.current?.querySelector('andypf-json-viewer')
      const shadowRoot = jsonViewer?.shadowRoot
      if (!shadowRoot) return

      const handler = (event: Event) => {
        const path = event.composedPath()
        const wrapper = path.find(
          (el): el is HTMLElement =>
            el instanceof HTMLElement &&
            el.classList.contains('icon-wrapper') &&
            !!el.querySelector(':scope > .copy.icon')
        )
        if (!wrapper) return

        const dataRow = wrapper.closest('.data-row')
        if (!dataRow) return

        const valueData = dataRow.querySelector(
          ':scope > .key-value-wrapper > .value.string > .value-data'
        )
        if (!valueData) return

        const text = valueData.textContent ?? ''
        const stripped =
          text.startsWith('"') && text.endsWith('"') && text.length >= 2
            ? text.slice(1, -1)
            : text

        navigator.clipboard.writeText(stripped)
        event.stopImmediatePropagation()
        event.preventDefault()
      }

      shadowRoot.addEventListener('click', handler, true)
      cleanup = () => shadowRoot.removeEventListener('click', handler, true)
    }

    const timer = setTimeout(attach, 100)

    return () => {
      clearTimeout(timer)
      cleanup?.()
    }
  }, [data])

  return (
    <div
      ref={containerRef}
      className={cn('border rounded-md overflow-auto', className)}
      {...props}
    >
      <Suspense fallback={<Skeleton height="450px" width="100%" />}>
        <JsonViewer
          key={colorScheme}
          data={data}
          theme={themeString}
          expanded={expanded}
          indent={indent}
          showDataTypes={showDataTypes}
          showToolbar={showToolbar}
          showCopy={showCopy}
          showSize={showSize}
          expandIconType={expandIconType}
        />
      </Suspense>
    </div>
  )
}

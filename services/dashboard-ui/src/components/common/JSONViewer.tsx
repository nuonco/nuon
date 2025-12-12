'use client'

import dynamic from 'next/dynamic'
import React, { useEffect, useRef } from 'react'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { cn } from '@/utils/classnames'

const JsonViewer = dynamic(
  () => import('@andypf/json-viewer/dist/esm/react/JsonViewer'),
  {
    loading: () => (
      <div className="border rounded-md overflow-auto p-4 animate-pulse">
        <div className="h-20 bg-cool-grey-200 dark:bg-dark-grey-700 rounded"></div>
      </div>
    ),
    ssr: false,
  }
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
  
  // Custom dark theme with correct JSONViewer Base16 mapping
  const customDarkTheme = {
    base00: '#19171C', // Default Background (bg-code dark)
    base01: '#19171C', // Lighter Background (unused but set same as base00)
    base02: '#3E4451', // Borders and Background for types NaN, null, undefined
    base03: '#5C6370', // Comments, Invisibles (unused but set for completeness)
    base04: '#9DA0A2', // Item Size (object/array sizes)
    base05: '#ABB2BF', // Default Foreground, Brackets, and Colons
    base06: '#ABB2BF', // Light Foreground (unused but set same as base05)
    base07: '#E06C75', // Keys, Colons, and Brackets (pink)
    base08: '#E06C75', // Color for NaN
    base09: '#98C379', // Ellipsis and String Values
    base0A: '#E5C07B', // Regular Expressions and Null Values
    base0B: '#61AFEF', // Floating-Point Values
    base0C: '#D19A66', // Number Keys
    base0D: '#56B6C2', // Icons, Search Input, Date
    base0E: '#C678DD', // Booleans and Expanded Icons
    base0F: '#D19A66', // Integers
  }

  // Custom light theme with correct JSONViewer Base16 mapping
  const customLightTheme = {
    base00: '#F0F3F5', // Default Background (bg-code light)
    base01: '#F0F3F5', // Lighter Background (unused but set same as base00)
    base02: '#E5E5E6', // Borders and Background for types NaN, null, undefined
    base03: '#A0A1A7', // Comments, Invisibles (unused but set for completeness)
    base04: '#696C77', // Item Size (object/array sizes)
    base05: '#383A42', // Default Foreground, Brackets, and Colons
    base06: '#383A42', // Light Foreground (unused but set same as base05)
    base07: '#E45649', // Keys, Colons, and Brackets (pink)
    base08: '#E45649', // Color for NaN
    base09: '#50A14F', // Ellipsis and String Values
    base0A: '#C18401', // Regular Expressions and Null Values
    base0B: '#4078F2', // Floating-Point Values
    base0C: '#C18401', // Number Keys
    base0D: '#0184BC', // Icons, Search Input, Date
    base0E: '#A626A4', // Booleans and Expanded Icons
    base0F: '#C18401', // Integers
  }

  const theme = colorScheme === 'dark' ? customDarkTheme : customLightTheme
  const themeString = JSON.stringify(theme)

  // Inject custom CSS into Shadow DOM
  useEffect(() => {
    const injectShadowCSS = () => {
      if (!containerRef.current) return
      
      const jsonViewer = containerRef.current.querySelector('andypf-json-viewer')
      if (!jsonViewer || !jsonViewer.shadowRoot) return
      
      // Create a new stylesheet
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
      
      // Add the stylesheet to the shadow root
      if (jsonViewer.shadowRoot.adoptedStyleSheets) {
        jsonViewer.shadowRoot.adoptedStyleSheets = [
          ...jsonViewer.shadowRoot.adoptedStyleSheets,
          sheet
        ]
      }
    }
    
    // Try injecting after a short delay to ensure the component is rendered
    const timer = setTimeout(injectShadowCSS, 100)
    
    return () => clearTimeout(timer)
  }, [data, colorScheme])

  return (
    <div
      ref={containerRef}
      className={cn(
        'border rounded-md overflow-auto',
        className
      )}
      {...props}
    >
      <style jsx>{`
        /* JSONViewer font styling */
        div :global(andypf-json-viewer),
        div :global(andypf-json-viewer) :global(.container),
        div :global(andypf-json-viewer) :global(.container) * {
          font-family: var(--font-hack) !important;
          font-size: 0.875rem !important;
          line-height: 1.25rem !important;
        }
        div :global(andypf-json-viewer) {
          /* Common JSON viewer CSS custom properties */
          --font-family: var(--font-hack);
          --font-size: 0.875rem;
          --line-height: 1.25rem;
          --json-font-family: var(--font-hack);
          --json-font-size: 0.875rem;
          --json-line-height: 1.25rem;
          --viewer-font-family: var(--font-hack);
          --viewer-font-size: 0.875rem;
          --text-font-size: 0.875rem;
          --code-font-family: var(--font-hack);
          --code-font-size: 0.875rem;
        }
      `}</style>
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
    </div>
  )
}

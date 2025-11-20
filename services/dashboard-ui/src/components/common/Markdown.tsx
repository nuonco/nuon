'use client'

import { useEffect, useState, useMemo } from 'react'
import showdown from 'showdown'
import { cn } from '@/utils/classnames'

showdown.extension('targetlink', () => {
  return [
    {
      type: 'output',
      regex: /<a\s+href="(?!#)(.*?)"(.*?)>/g,
      replace: '<a href="$1" target="_blank" $2>',
    },
  ]
})

showdown.extension('wrapTables', () => [
  {
    type: 'output',
    filter: (text: string) =>
      text.replace(
        /(<table[^>]*>[\s\S]*?<\/table>)/g,
        '<div class="readme-table">$1</div>'
      ),
  },
])

let mermaidCounter = 0

showdown.extension('mermaid', () => [
  {
    type: 'output',
    filter: (text: string) => {
      return text.replace(
        /<pre><code class="mermaid language-mermaid">([\s\S]*?)<\/code><\/pre>/g,
        (match, code) => {
          const id = `mermaid-${++mermaidCounter}`
          // Properly decode HTML entities and preserve br tags
          const decodedCode = code
            .replace(/&lt;/g, '<')
            .replace(/&gt;/g, '>')
            .replace(/&amp;/g, '&')
            // Keep HTML br tags - they work in mermaid node labels
            .replace(/&lt;br\/&gt;/g, '<br/>')
            .replace(/&lt;br&gt;/g, '<br>')
          return `<div class="mermaid-diagram" id="${id}" data-mermaid="${encodeURIComponent(decodedCode.trim())}"></div>`
        }
      )
    },
  },
])

const markdown = new showdown.Converter({
  extensions: ['targetlink', 'wrapTables', 'mermaid'],
  tables: true,
  tasklists: true,
})

export const Markdown = ({ content = '' }) => {
  const [processedHtml, setProcessedHtml] = useState('')
  const [renderedSvgs, setRenderedSvgs] = useState<Record<string, string>>({})

  // Generate HTML from markdown
  const baseHtml = useMemo(() => {
    return markdown.makeHtml(content)
  }, [content])

  // Process mermaid diagrams and update HTML
  useEffect(() => {
    const processMermaid = async () => {
      let html = baseHtml
      const mermaidRegex =
        /<div class="mermaid-diagram" id="([^"]+)" data-mermaid="([^"]+)"><\/div>/g
      let match
      const newSvgs = { ...renderedSvgs }
      let hasNewSvgs = false

      while ((match = mermaidRegex.exec(baseHtml)) !== null) {
        const [fullMatch, id, encodedDefinition] = match

        // Skip if already rendered
        if (renderedSvgs[id]) {
          html = html.replace(
            fullMatch,
            `<div class="mermaid-diagram" id="${id}">${renderedSvgs[id]}</div>`
          )
          continue
        }

        hasNewSvgs = true
        try {
          const definition = decodeURIComponent(encodedDefinition)
          const mermaid = await import('mermaid')

          mermaid.default.initialize({
            startOnLoad: false,
            theme: 'default',
            securityLevel: 'loose',
            fontFamily: 'inherit',
          })

          const { svg } = await mermaid.default.render(
            `${id}-render`,
            definition
          )
          newSvgs[id] = svg
          html = html.replace(
            fullMatch,
            `<div class="mermaid-diagram" id="${id}">${svg}</div>`
          )
        } catch (error) {
          console.error('Mermaid error:', error)
          const errorHtml = `<pre class="text-red-600 dark:text-red-400">Mermaid Error: ${error}</pre>`
          newSvgs[id] = errorHtml
          html = html.replace(
            fullMatch,
            `<div class="mermaid-diagram" id="${id}">${errorHtml}</div>`
          )
        }
      }

      // Only update state if there are new SVGs
      if (hasNewSvgs) {
        setRenderedSvgs(newSvgs)
      }
      setProcessedHtml(html)
    }

    processMermaid()
  }, [content])

  return (
    <>
      <style>{`
        .prose .readme-table pre { max-width: 50ch; } 
        .mermaid-diagram { 
          text-align: center; 
          margin: 1rem 0; 
          min-height: 100px;
          border-radius: 0.25rem;
          border: 1px solid var(--border-color);
          padding: 1rem;
          background: var(--background-neutral);
        }
        .mermaid-diagram svg {
          max-width: 100%;
          height: 300px;
        }
      `}</style>
      <div
        className={cn(
          'prose dark:prose-invert max-w-[100%]',
          'prose-code:bg-code prose-code:text-sm prose-code:text-blue-500 prose-code:font-mono',
          'prose-pre:bg-code prose-pre:text-sm prose-pre:text-blue-500 prose-pre:font-mono prose-pre:rounded prose-pre:shadow-sm prose-pre:overflow-auto prose-pre:max-w-[80ch]'
        )}
        dangerouslySetInnerHTML={{
          __html: processedHtml || baseHtml,
        }}
      />
    </>
  )
}

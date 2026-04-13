import React, { useEffect, useId, useMemo, useState, lazy, Suspense, type ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { cn } from '@/utils/classnames'
import { CodeBlock } from './CodeBlock'
import { JSONViewer } from './JSONViewer'
import { Link } from './Link'
import { Tabs } from './Tabs'
import { Button } from './Button'
import { ModalBase } from '@/components/surfaces/Modal'
import { PanelBase } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { buildNuonComponents, extractTabs, extractSurfaces, nuonTagNames, type ExtractedTabs, type ExtractedSurface, type MarkdownMode } from './markdown-components'

const MermaidFlowGraph = lazy(() => import('./MermaidFlowGraph').then((m) => ({ default: m.MermaidFlowGraph })))

const BLOCK_TAG_NAMES = new Set([...nuonTagNames, 'nuon-surface-rendered'])

const REMARK_PLUGINS = [remarkGfm] as const
const REHYPE_PLUGINS = [rehypeRaw] as const

const isFlowchart = (code: string) => /^(?:graph|flowchart)\s+(?:TD|TB|LR|RL|BT)\s*$/im.test(code.trim().split('\n')[0])

const MermaidSvgDiagram = ({ code }: { code: string }) => {
  const [svg, setSvg] = useState<string>('')
  const [error, setError] = useState<string>('')
  const reactId = useId()
  const id = `mermaid-${reactId.replace(/:/g, '')}`

  useEffect(() => {
    const renderMermaid = async () => {
      try {
        const mermaid = await import('mermaid')
        mermaid.default.initialize({
          startOnLoad: false,
          theme: 'neutral',
          securityLevel: 'loose',
          fontFamily: 'inherit',
        })

        const { svg: renderedSvg } = await mermaid.default.render(`${id}-render`, code)
        const cleaned = renderedSvg
          .replace(/\s+height="[\d.]+"/, '')
          .replace(/\s+style="[^"]*max-width:[^"]*"/, ' style="width: 100%"')
        setSvg(cleaned)
      } catch (err) {
        console.error('Mermaid error:', err)
        setError(`Mermaid Error: ${err}`)
      }
    }

    renderMermaid()
  }, [code, id])

  if (error) {
    return <pre className="text-red-600 dark:text-red-400">{error}</pre>
  }

  return (
    <div
      className="mermaid-diagram text-center my-4 min-h-[100px] border rounded-lg border-color p-4 bg-white dark:invert dark:hue-rotate-180"
      dangerouslySetInnerHTML={{ __html: svg || 'Loading diagram...' }}
    />
  )
}

function renderCodeBlock(language: string, codeString: string) {
  if (language === 'mermaid') {
    if (isFlowchart(codeString)) {
      return (
        <Suspense fallback={<div className="w-full h-[44rem] my-4 border rounded-lg border-color animate-pulse" />}>
          <MermaidFlowGraph code={codeString} />
        </Suspense>
      )
    }
    return <MermaidSvgDiagram code={codeString} />
  }

  if (language === 'json' || language === 'jsonc') {
    try {
      return <JSONViewer data={JSON.parse(codeString)} expanded={2} className="my-4" />
    } catch {
      return <CodeBlock language="json">{codeString}</CodeBlock>
    }
  }

  if (!language || language === 'text') {
    try {
      return <JSONViewer data={JSON.parse(codeString)} expanded={2} className="my-4" />
    } catch {
      return (
        <pre className="overflow-x-auto rounded-lg border p-4 my-4 bg-code text-sm font-mono">
          <code>{codeString}</code>
        </pre>
      )
    }
  }

  return <CodeBlock language={language}>{codeString}</CodeBlock>
}

const nuonComponentsByMode: Record<MarkdownMode, Record<string, any>> = {
  app: buildNuonComponents('app'),
  install: buildNuonComponents('install'),
}

function hasBlockChild(node: any): boolean {
  return node?.children?.some(
    (child: any) => child.type === 'element' && BLOCK_TAG_NAMES.has(child.tagName)
  )
}

function getMarkdownComponents(mode: MarkdownMode): Record<string, any> {
  return {
  ...nuonComponentsByMode[mode],
  p({ node, children, ...props }: any) {
    if (hasBlockChild(node)) {
      return <div {...props}>{children}</div>
    }
    return <p {...props}>{children}</p>
  },
  code({ node, className, children, style, ...props }: any) {
      if (style || node?.properties?.style) {
        return <code className={className} style={style} {...props}>{children}</code>
      }

      return (
        <code
          className={cn(
            'bg-code text-sm text-blue-800 dark:text-blue-500 font-mono px-1 py-0.5 rounded',
            className
          )}
          {...props}
        >
          {children}
        </code>
      )
    },
    
    // Handle links using the custom Link component
    a({ href, children, ...props }: any) {
      const isExternal = href && !href.startsWith('#') && !href.startsWith('/')
      return (
        <Link 
          href={href}
          isExternal={isExternal}
          {...props}
        >
          {children}
        </Link>
      )
    },
    
    table({ children, ...props }: any) {
      return (
        <div className="readme-table overflow-x-auto rounded-lg border my-4">
          <table className="min-w-full text-sm !my-0" {...props}>{children}</table>
        </div>
      )
    },
    thead({ children, ...props }: any) {
      return <thead {...props}>{children}</thead>
    },
    th({ children, ...props }: any) {
      return (
        <th className="py-3 px-4 text-left bg-cool-grey-100 dark:bg-dark-grey-700 first:rounded-tl-lg last:rounded-tr-lg" {...props}>
          {children}
        </th>
      )
    },
    td({ children, ...props }: any) {
      return (
        <td className="py-3 px-4 border-t" {...props}>
          {children}
        </td>
      )
    },
    
    // Handle details/summary for collapsible content using styled approach
    details({ children, ...props }: any) {
      // Separate summary from other content
      const childrenArray = React.Children.toArray(children)
      const summaryChild = childrenArray.find((child: any) => child?.type === 'summary' || child?.props?.node?.tagName === 'summary')
      const contentChildren = childrenArray.filter((child: any) => child?.type !== 'summary' && child?.props?.node?.tagName !== 'summary')
      
      return (
        <details 
          className="expand-wrapper shrink-0 flex flex-col w-full overflow-hidden my-4 border rounded-lg"
          {...props}
        >
          {summaryChild}
          {contentChildren.length > 0 && (
            <div className="p-4">
              {contentChildren}
            </div>
          )}
        </details>
      )
    },
    
    summary({ children, ...props }: any) {
      return (
        <summary 
          className="flex items-center gap-2 cursor-pointer px-4 py-3 w-full outline-none transition-all hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10 list-none [&::-webkit-details-marker]:hidden overflow-hidden"
          {...props}
        >
          <span className="flex-1">{children}</span>
          <svg 
            className="w-4 h-4 transition-transform duration-200 [details[open]>&]:rotate-180" 
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </summary>
      )
    },

  pre({ node, children, style, ...props }: any) {
    if (style || node?.properties?.style) {
      return <pre style={style} {...props}>{children}</pre>
    }

    const childArray = React.Children.toArray(children)
    const child = childArray.length === 1 ? (childArray[0] as React.ReactElement<any>) : null

    if (child?.props?.className) {
      const match = /language-(\w+)/.exec(child.props.className)
      if (match) {
        const codeString = String(child.props.children).replace(/\n$/, '')
        return renderCodeBlock(match[1], codeString)
      }
    }

    if (child?.props?.children != null) {
      const codeString = String(child.props.children).replace(/\n$/, '')
      return renderCodeBlock('text', codeString)
    }

    return <pre style={style} {...props}>{children}</pre>
  },
  }
}

function preprocessContent(content: string): string {
  const lines = content.split('\n')
  const result: string[] = []
  let htmlDepth = 0

  for (const line of lines) {
    const opens = (line.match(/<(?:div|table|thead|tbody|tr|ul|ol|section)\b/gi) || []).length
    const closes = (line.match(/<\/(?:div|table|thead|tbody|tr|ul|ol|section)\b/gi) || []).length
    htmlDepth += opens - closes

    if (htmlDepth > 0 && line.trim() === '') {
      continue
    }

    result.push(line)

    if (htmlDepth < 0) htmlDepth = 0
  }

  return result.join('\n')
}

const markdownStyles = `
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
    height: auto;
  }
  details[open] > summary {
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 0;
  }
  details div ul, details div ol {
    padding-left: 1.5rem;
  }
  .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::before,
  .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::after,
  .prose code::before,
  .prose code::after {
    content: none !important;
  }
  .prose andypf-json-viewer *[class*="container"],
  .prose andypf-json-viewer *[class="container"],
  .prose andypf-json-viewer div.container,
  .prose andypf-json-viewer div.container *,
  .prose div andypf-json-viewer,
  .prose div andypf-json-viewer .container,
  .prose div andypf-json-viewer .container * {
    font-family: var(--font-hack) !important;
    font-size: 0.875rem !important;
    line-height: 1.25rem !important;
  }
  andypf-json-viewer .container,
  andypf-json-viewer .container * {
    font-family: var(--font-hack) !important;
    font-size: 0.875rem !important;
    line-height: 1.25rem !important;
  }
  .prose andypf-json-viewer {
    --font-family: var(--font-hack) !important;
    --font-size: 0.875rem !important;
    --line-height: 1.25rem !important;
    --json-font-family: var(--font-hack) !important;
    --json-font-size: 0.875rem !important;
    --json-line-height: 1.25rem !important;
    --viewer-font-family: var(--font-hack) !important;
    --viewer-font-size: 0.875rem !important;
    --text-font-size: 0.875rem !important;
    --code-font-family: var(--font-hack) !important;
    --code-font-size: 0.875rem !important;
    --container-font-family: var(--font-hack) !important;
    --container-font-size: 0.875rem !important;
    --key-value-font-family: var(--font-hack) !important;
    --key-value-font-size: 0.875rem !important;
    --value-string-font-family: var(--font-hack) !important;
    --value-string-font-size: 0.875rem !important;
  }
`

const proseClassName = cn(
  'prose dark:prose-invert max-w-[100%]',
  'prose-code:bg-code prose-code:text-sm prose-code:text-blue-500 prose-code:font-mono'
)

function TabsPlaceholder({
  tabsMap,
  mode,
  dataId,
}: {
  tabsMap: Map<string, ExtractedTabs>
  mode: MarkdownMode
  dataId: string
}) {
  const extracted = tabsMap.get(dataId)
  if (!extracted) return null

  const tabs: Record<string, ReactNode> = {}
  for (const tab of extracted) {
    tabs[tab.name] = <Markdown content={tab.content} mode={mode} />
  }
  return <Tabs tabs={tabs} />
}

function SurfacePlaceholder({
  surfaceMap,
  mode,
  dataId,
}: {
  surfaceMap: Map<string, ExtractedSurface>
  mode: MarkdownMode
  dataId: string
}) {
  const { addModal, addPanel } = useSurfaces()
  const surface = surfaceMap.get(dataId)
  if (!surface) return null

  const handleClick = () => {
    const content = <Markdown content={surface.content} mode={mode} />
    if (surface.type === 'modal') {
      addModal(
        <ModalBase heading={surface.heading} size={surface.size as any} showFooter={false}>
          {content}
        </ModalBase>
      )
    } else {
      addPanel(
        <PanelBase heading={surface.heading} size={surface.size as any}>
          {content}
        </PanelBase>
      )
    }
  }

  return (
    <Button variant="secondary" onClick={handleClick}>
      {surface.trigger}
    </Button>
  )
}

export const Markdown = React.memo(({ content = '', mode = 'app' }: { content?: string; mode?: MarkdownMode }) => {
  const { content: processedContent, tabsMap, surfaceMap } = useMemo(() => {
    const { content: afterTabs, tabsMap } = extractTabs(content)
    const { content: afterSurfaces, surfaceMap } = extractSurfaces(afterTabs)
    return { content: afterSurfaces, tabsMap, surfaceMap }
  }, [content])
  const processed = preprocessContent(processedContent)

  const components = useMemo(() => {
    const base = getMarkdownComponents(mode)
    if (tabsMap.size > 0) {
      base['nuon-tabs-rendered'] = ({ node, ...attrs }: any) => (
        <TabsPlaceholder tabsMap={tabsMap} mode={mode} dataId={attrs['data-id']} />
      )
    }
    if (surfaceMap.size > 0) {
      base['nuon-surface-rendered'] = ({ node, ...attrs }: any) => (
        <SurfacePlaceholder surfaceMap={surfaceMap} mode={mode} dataId={attrs['data-id']} />
      )
    }
    return base
  }, [mode, tabsMap, surfaceMap])

  return (
    <>
      <style>{markdownStyles}</style>
      <div className={proseClassName}>
        <ReactMarkdown
          remarkPlugins={REMARK_PLUGINS}
          rehypePlugins={REHYPE_PLUGINS}
          components={components}
        >
          {processed}
        </ReactMarkdown>
      </div>
    </>
  )
})

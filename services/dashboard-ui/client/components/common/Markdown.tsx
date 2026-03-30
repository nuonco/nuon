import React, { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { cn } from '@/utils/classnames'
import { CodeBlock } from './CodeBlock'
import { JSONViewer } from './JSONViewer'
import { Link } from './Link'

// Mermaid component that handles its own rendering
const MermaidDiagram = ({ code }: { code: string }) => {
  const [svg, setSvg] = useState<string>('')
  const [error, setError] = useState<string>('')
  const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`

  useEffect(() => {
    const renderMermaid = async () => {
      try {
        const mermaid = await import('mermaid')
        mermaid.default.initialize({
          startOnLoad: false,
          theme: 'default',
          securityLevel: 'loose',
          fontFamily: 'inherit',
        })
        
        const { svg: renderedSvg } = await mermaid.default.render(`${id}-render`, code)
        setSvg(renderedSvg)
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
      className="mermaid-diagram text-center my-4 min-h-[100px] border rounded border-color p-4 bg-neutral"
      dangerouslySetInnerHTML={{ __html: svg || 'Loading diagram...' }}
    />
  )
}

function renderCodeBlock(language: string, codeString: string) {
  if (language === 'mermaid') {
    return <MermaidDiagram code={codeString} />
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

const markdownComponents = {
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

export const Markdown = React.memo(({ content = '' }: { content?: string }) => {
  const processed = preprocessContent(content)
  return (
    <>
      <style>{`
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
        /* Enhanced details/summary styling to match Expand component */
        details[open] > summary {
          border-bottom: 1px solid var(--border-color);
          margin-bottom: 0;
        }
        /* Fix list padding inside details content wrapper */
        details div ul, details div ol {
          padding-left: 1.5rem;
        }
        /* Remove any pseudo-element backticks from inline code */
        .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::before,
        .prose :where(code):not(:where([class~="not-prose"], [class~="not-prose"] *))::after,
        .prose code::before,
        .prose code::after {
          content: none !important;
        }
        /* Override prose styles for JSONViewer custom element - nuclear approach */
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
        /* Global override with maximum specificity */
        andypf-json-viewer .container,
        andypf-json-viewer .container * {
          font-family: var(--font-hack) !important;
          font-size: 0.875rem !important;
          line-height: 1.25rem !important;
        }
        .prose andypf-json-viewer {
          /* Shadow DOM CSS custom properties - try various naming patterns */
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
          /* Try container-specific variables */
          --container-font-family: var(--font-hack) !important;
          --container-font-size: 0.875rem !important;
          /* Try more specific variables based on classes we see */
          --key-value-font-family: var(--font-hack) !important;
          --key-value-font-size: 0.875rem !important;
          --value-string-font-family: var(--font-hack) !important;
          --value-string-font-size: 0.875rem !important;
        }
      `}</style>
      <div className={cn(
        'prose dark:prose-invert max-w-[100%]',
        'prose-code:bg-code prose-code:text-sm prose-code:text-blue-500 prose-code:font-mono'
      )}>
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeRaw]}
          components={markdownComponents}
        >
          {processed}
        </ReactMarkdown>
      </div>
    </>
  )
})

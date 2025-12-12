'use client'

import React, { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { cn } from '@/utils/classnames'
import { CodeBlock } from './CodeBlock'
import { JSONViewer } from './JSONViewer'
import { Link } from './Link'
import { Expand } from './Expand'
import { Code } from './Code'

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

export const Markdown = ({ content = '' }) => {
  // Custom components for react-markdown
  const components = {
    // Handle code blocks
    code({ node, inline, className, children, ...props }: any) {
      const match = /language-(\w+)/.exec(className || '')
      const language = match ? match[1] : 'text'
      const codeString = String(children).replace(/\n$/, '')
      
      // Detect inline code: explicit inline prop or no className (no language specified)
      const isInlineCode = inline === true || !className
      
      if (!isInlineCode) {
        // Block code
        
        // Handle mermaid diagrams
        if (language === 'mermaid') {
          return <MermaidDiagram code={codeString} />
        }
        
        // Handle JSON
        if (language === 'json' || language === 'jsonc') {
          try {
            const jsonData = JSON.parse(codeString)
            return <JSONViewer data={jsonData} expanded={2} className="my-4" />
          } catch {
            // Fallback to regular code block if JSON parsing fails
            return <CodeBlock language="json">{codeString}</CodeBlock>
          }
        }
        
        // Auto-detect JSON if no language specified
        if (language === 'text' || !language) {
          try {
            const jsonData = JSON.parse(codeString)
            return <JSONViewer data={jsonData} expanded={2} className="my-4" />
          } catch {
            // Not JSON, use regular code block
          }
        }
        
        // Regular code block
        return <CodeBlock language={language}>{codeString}</CodeBlock>
      }
      
      // Inline code
      return (
        <code 
          className={cn(
            'bg-code text-sm text-blue-800 dark:text-blue-500 font-mono px-1 py-0.5 rounded',
            className
          )} 
          {...props}
          style={{ position: 'relative' }}
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
    
    // Handle tables (wrap in container for styling)
    table({ children, ...props }: any) {
      return (
        <div className="readme-table">
          <table {...props}>{children}</table>
        </div>
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

    // Override pre to prevent double wrapping of CodeBlock and JSONViewer
    pre({ children, ...props }: any) {
      // Just return the children directly to avoid wrapping CodeBlock/JSONViewer
      return <>{children}</>
    },
  }

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
        .prose code:before,
        .prose code:after {
          content: '' !important;
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
          components={components}
        >
          {content}
        </ReactMarkdown>
      </div>
    </>
  )
}

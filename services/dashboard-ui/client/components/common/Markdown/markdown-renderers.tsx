import React, { lazy, Suspense } from 'react'
import { cn } from '@/utils/classnames'
import { CodeBlock } from '../CodeBlock'
import { JSONViewer } from '../JSONViewer'
import { Link } from '../Link'
import { buildNuonComponents, nuonTagNames, type MarkdownMode } from './nuon-components'

const MermaidFlowGraph = lazy(() => import('../MermaidFlowGraph').then((m) => ({ default: m.MermaidFlowGraph })))

const BLOCK_TAG_NAMES = new Set([...nuonTagNames, 'nuon-surface-rendered'])

const isFlowchart = (code: string) => /^(?:graph|flowchart)\s+(?:TD|TB|LR|RL|BT)\s*$/im.test(code.trim().split('\n')[0])

function renderCodeBlock(language: string, codeString: string) {
  if (language === 'mermaid') {
    if (isFlowchart(codeString)) {
      return (
        <Suspense fallback={<div className="w-full h-[44rem] my-4 border rounded-lg border-color animate-pulse" />}>
          <MermaidFlowGraph code={codeString} />
        </Suspense>
      )
    }
    return <CodeBlock language="mermaid">{codeString}</CodeBlock>
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

export function getMarkdownComponents(mode: MarkdownMode): Record<string, any> {
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

    details({ children, ...props }: any) {
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

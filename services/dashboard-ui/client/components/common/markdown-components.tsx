import type { ComponentType } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Status } from '@/components/common/Status'
import { InstallConfigGraph } from '@/components/installs/InstallConfigGraph'
import { ViewStateButton } from '@/components/installs/management/ViewState'

export type MarkdownMode = 'app' | 'install'

type NuonComponent = {
  component: ComponentType<any>
  mapProps: (attrs: Record<string, string>) => Record<string, any>
  requiresInstall?: boolean
}

const noProps = () => ({})

const registry: Record<string, NuonComponent> = {
  'nuon-badge': {
    component: Badge,
    mapProps: (attrs) => ({
      children: attrs.children,
      theme: attrs.theme,
      size: attrs.size,
      variant: attrs.variant,
    }),
  },
  'nuon-banner': {
    component: Banner,
    mapProps: (attrs) => ({
      children: attrs.children,
      theme: attrs.theme,
    }),
  },
  'nuon-status': {
    component: Status,
    mapProps: (attrs) => ({
      status: attrs.status,
      variant: attrs.variant,
    }),
  },
  'nuon-view-state': {
    component: ViewStateButton,
    mapProps: noProps,
    requiresInstall: true,
  },
  'nuon-config-graph': {
    component: InstallConfigGraph,
    mapProps: noProps,
    requiresInstall: true,
  },
}

export type ExtractedTabs = { name: string; content: string }[]

export type ExtractedSurface = {
  type: 'modal' | 'panel'
  heading: string
  size: string
  trigger: string
  content: string
}

function parseTagAttrs(tag: string): Record<string, string> {
  const attrs: Record<string, string> = {}
  const attrRegex = /(\w+)="([^"]*)"/g
  let m: RegExpExecArray | null
  while ((m = attrRegex.exec(tag)) !== null) {
    attrs[m[1]] = m[2]
  }
  return attrs
}

export function extractTabs(content: string): {
  content: string
  tabsMap: Map<string, ExtractedTabs>
} {
  const tabsMap = new Map<string, ExtractedTabs>()
  let idx = 0

  const replaced = content.replace(
    /<nuon-tabs>([\s\S]*?)<\/nuon-tabs>/g,
    (_match, inner: string) => {
      const tabRegex = /<nuon-tab\s+name="([^"]+)">([\s\S]*?)<\/nuon-tab>/g
      const tabs: ExtractedTabs = []
      let tabMatch: RegExpExecArray | null
      while ((tabMatch = tabRegex.exec(inner)) !== null) {
        tabs.push({ name: tabMatch[1], content: tabMatch[2].trim() })
      }
      if (tabs.length === 0) return ''
      const id = `nuon-tabs-${idx++}`
      tabsMap.set(id, tabs)
      return `<nuon-tabs-rendered data-id="${id}"></nuon-tabs-rendered>`
    }
  )

  return { content: replaced, tabsMap }
}

export function extractSurfaces(content: string): {
  content: string
  surfaceMap: Map<string, ExtractedSurface>
} {
  const surfaceMap = new Map<string, ExtractedSurface>()
  let idx = 0

  const replaced = content.replace(
    /<nuon-(modal|panel)\s([^>]*)>([\s\S]*?)<\/nuon-\1>/g,
    (_match, type: string, attrStr: string, inner: string) => {
      const attrs = parseTagAttrs(attrStr)
      const id = `nuon-surface-${idx++}`
      surfaceMap.set(id, {
        type: type as 'modal' | 'panel',
        heading: attrs.heading || '',
        size: attrs.size || 'default',
        trigger: attrs.trigger || 'View',
        content: inner.trim(),
      })
      return `<nuon-surface-rendered data-id="${id}"></nuon-surface-rendered>`
    }
  )

  return { content: replaced, surfaceMap }
}

function hasUnresolvedTemplates(attrs: Record<string, string>): boolean {
  return Object.values(attrs).some(
    (v) => typeof v === 'string' && v.includes('{{')
  )
}

function InlineCodeFallback({ tagName, attrs }: { tagName: string; attrs: Record<string, string> }) {
  const attrStr = Object.entries(attrs)
    .filter(([k]) => k !== 'children')
    .map(([k, v]) => `${k}="${v}"`)
    .join(' ')
  const tag = attrStr ? `<${tagName} ${attrStr}>` : `<${tagName}>`
  const children = attrs.children
  const display = children ? `${tag}${children}</${tagName}>` : `${tag}</${tagName}>`
  return (
    <code className="bg-code text-sm text-blue-800 dark:text-blue-500 font-mono px-1 py-0.5 rounded">
      {display}
    </code>
  )
}

export function buildNuonComponents(
  mode: MarkdownMode = 'app'
): Record<string, ComponentType<any>> {
  const components: Record<string, ComponentType<any>> = {}

  for (const [tagName, { component: Component, mapProps, requiresInstall }] of Object.entries(
    registry
  )) {
    components[tagName] = ({ node, ...attrs }: any) => {
      if (requiresInstall && mode === 'app') {
        return <InlineCodeFallback tagName={tagName} attrs={attrs} />
      }

      if (!requiresInstall && hasUnresolvedTemplates(attrs)) {
        return <InlineCodeFallback tagName={tagName} attrs={attrs} />
      }

      const props = mapProps(attrs)
      return <Component {...props} />
    }
  }

  return components
}

export const nuonTagNames = Object.keys(registry)

import React, { useMemo, type ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { Tabs } from '../Tabs'
import { Button } from '../Button'
import { ModalBase } from '@/components/surfaces/Modal'
import { PanelBase } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { extractTabs, extractSurfaces, type ExtractedTabs, type ExtractedSurface, type MarkdownMode } from './nuon-components'
import { getMarkdownComponents } from './markdown-renderers'
import { preprocessContent } from './markdown-preprocessing'
import { markdownStyles, proseClassName } from './markdown-styles'

const REMARK_PLUGINS = [remarkGfm] as const
const REHYPE_PLUGINS = [rehypeRaw] as const

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

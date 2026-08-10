import type { ReactNode } from 'react'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Tabs } from '@/components/common/Tabs'

export const TerraformRenderedVariablesFiles = ({
  files,
}: {
  files: string[]
}) => {
  if (files.length === 1) {
    return (
      <CodeBlock language="hcl" className="!max-h-fit">
        {files[0]}
      </CodeBlock>
    )
  }

  const fileTabs: Record<string, ReactNode> = {}
  files.forEach((file, idx) => {
    fileTabs[`file ${idx + 1}`] = (
      <CodeBlock language="hcl" className="!max-h-fit">
        {file}
      </CodeBlock>
    )
  })

  return <Tabs tabs={fileTabs} />
}

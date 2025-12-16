'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TSandboxRun } from '@/types'

interface ISandboxRunOutputsModal {
  sandboxRun: TSandboxRun
  headingText?: string
}

export const SandboxRunOutputsModal = ({
  sandboxRun,
  headingText = 'Sandbox run outputs',
  ...props
}: ISandboxRunOutputsModal & IModal) => {
  const outputs = sandboxRun?.outputs

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="CodeBlock" size="24" />
          {headingText}
        </Text>
      }
      className="!max-w-5xl"
      {...props}
    >
      <div className="flex flex-col gap-4">
        {!outputs ? (
          <Banner theme="info">
            No outputs available for this sandbox run.
          </Banner>
        ) : Object.keys(outputs).length === 0 ? (
          <div className="flex items-center justify-center p-8">
            <Text variant="body" theme="neutral">
              No output data available
            </Text>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <ClickToCopyButton
                textToCopy={JSON.stringify(outputs, null, 2)}
                className="w-fit"
              />
            </div>
            <CodeBlock language="json" className="max-h-[600px]">
              {JSON.stringify(outputs, null, 2)}
            </CodeBlock>
          </div>
        )}
      </div>
    </Modal>
  )
}

export const SandboxRunOutputsButton = ({
  sandboxRun,
  headingText,
  ...props
}: ISandboxRunOutputsModal & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <SandboxRunOutputsModal sandboxRun={sandboxRun} headingText={headingText} />
  )

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CodeBlock" />}
      View outputs
      {props?.isMenuButton ? <Icon variant="CodeBlock" /> : null}
    </Button>
  )
}
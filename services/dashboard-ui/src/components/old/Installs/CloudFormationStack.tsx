import React, { type FC } from 'react'
import { ClickToCopyButton } from '@/components/old/ClickToCopy'
import { ToolTip } from '@/components/old/ToolTip'
import { Text, Code } from '@/components/old/Typography'

export const CloudFormationLink: FC<{ cfLink: string }> = ({
  cfLink = 'https://us-east-1.console.aws.amazon.com/cloudformation/home?region=us-east-1#',
}) => {
  return (
    <div className="border rounded-lg p-3 flex flex-col gap-3">
      <div className="flex justify-between items-center">
        <Text variant="med-14">
          <ToolTip tipContent="Not sure what needs to go here">
            CloudFormation Stack
          </ToolTip>
        </Text>
        <ClickToCopyButton
          className="border rounded-md p-1 text-sm"
          textToCopy={cfLink}
        />
      </div>
      <Code>{cfLink}</Code>
    </div>
  )
}

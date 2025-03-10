import React, { type FC } from 'react'
import { Link } from '@/components/Link'
import { Text, CodeInline } from '@/components/Typography'

export const NoActions: FC = () => {
  return (
    <div className="max-w-xl flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Text variant="semi-18">No actions configured</Text>
        <Text className="!text-lg !leading-loose">
          Add action workflows to your app to run diagnostics, post install jobs
          and more.
          <br />
          <Link
            className="!inline"
            href="https://docs.nuon.co/concepts/nuon-actions"
            target="_blank"
          >
            Learn more
          </Link>{' '}
          about actions.
        </Text>
      </div>

      {/* <div className="flex flex-col gap-2">
          <Text variant="semi-14">Add an action to an app with the Nuon CLI</Text>
          <Text>
          Create your first action using the Nuon CLI with the{' '}
          <CodeInline variant="inline">
          nuon action create -n YOUR-APP-NAME
          </CodeInline>{' '}
          command. Follow our{' '}
          <Link
          href="https://docs.nuon.co/tutorials/aws-ecs-app-tutorial"
          target="_blank"
          >
          app creation guide
          </Link>{' '}
          to get started.
          </Text>
          </div> */}
    </div>
  )
}

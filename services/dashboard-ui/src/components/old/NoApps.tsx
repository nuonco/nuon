import React, { type FC } from 'react'
import { Link } from '@/components/old/Link'
import { Text, CodeInline } from '@/components/old/Typography'

export const NoApps: FC = () => {
  return (
    <div className="max-w-xl flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Text variant="semi-18">No apps configured</Text>
        <Text className="!text-lg !leading-loose">
          Package your existing application code and infrastructure to create
          bring-your-own-cloud installable versions of your product. Nuon apps
          are versions of your application code and infrastructure that can be
          deployed into a customer cloud account. <br />
          <Link
            className="!inline"
            href="https://docs.nuon.co/concepts/apps"
            target="_blank"
          >
            Learn more
          </Link>{' '}
          about apps.
        </Text>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="semi-14">Create an app with the Nuon CLI</Text>
        <Text>
          Create your first app using the Nuon CLI with the{' '}
          <CodeInline variant="inline">
            nuon apps create -n YOUR-APP-NAME
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
      </div>
    </div>
  )
}

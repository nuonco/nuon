import React, { type FC } from 'react'
import { CreateInstallModal } from '@/components/old/Installs'
import { Link } from '@/components/old/Link'
import { Text, CodeInline } from '@/components/old/Typography'

export const NoInstalls: FC = () => {
  return (
    <div className="max-w-xl flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Text variant="semi-18">No installs created</Text>
        <Text className="!text-lg !leading-loose">
          An install is an instance of an application running in a customer
          cloud account. When an install is created, each component in your
          application will be deployed in the correct order, alongside a runner
          which is responsible for managing the software. <br />
          <Link
            className="!inline"
            href="https://docs.nuon.co/concepts/installs"
            target="_blank"
          >
            Learn more
          </Link>{' '}
          about installs.
        </Text>
      </div>

      <CreateInstallModal />

      <div className="flex flex-col gap-2">
        <Text variant="semi-14">Create an install with the Nuon CLI</Text>
        <Text>
          If you&apos;ve already setup an AWS account with Nuon access create an
          install with the{' '}
          <CodeInline>
            nuon installs create -n YOUR-INSTALL-NAME -r AWS-REGION -o
            ARN-AWS-IAM
          </CodeInline>{' '}
          command. Or follow our{' '}
          <Link
            href="https://docs.nuon.co/guides/install-access-permissions#install-access-permissions"
            target="_blank"
          >
            install access permissions guide
          </Link>{' '}
          to get started.
        </Text>
      </div>
    </div>
  )
}

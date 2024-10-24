import React, { type FC } from 'react'
import { Card } from '@/components/Card'
import { Link } from '@/components/Link'
import { Heading, Text } from '@/components/Typography'

export const NoComponents: FC = () => {
  return (
    <div className="max-w-xl flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Heading variant="title">No components configured</Heading>
        <Text className="!text-lg !leading-loose">
          Connect and configure your application using your existing container
          images, infrastructure as code and devops automation. Nuon supports
          several different component types which allow you to connect your
          existing devops automation, infrastructure as code and containers to
          your app. <br />
          <Link
            className="!inline"
            href="https://docs.nuon.co/concepts/components"
            target="_blank"
          >
            Learn more
          </Link>{' '}
          about components.
        </Text>
      </div>

      <div className="flex flex-col gap-2">
        <Heading>Add components an app config</Heading>

        <div className="grid grid-cols-2 gap-4">
          <Card>
            <Text className="!font-medium">Docker component</Text>
            <Text variant="caption">Any Dockerfile that can be built</Text>
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/docker-build-components"
              target="_blank"
            >
              Learn more
            </Link>
          </Card>
          <Card>
            <Text className="!font-medium">Container image component</Text>
            <Text variant="caption">Any prebuilt OCI image</Text>
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/container-image-components"
              target="_blank"
            >
              Learn more
            </Link>
          </Card>
          <Card>
            <Text className="!font-medium">Helm chart component</Text>
            <Text variant="caption">
              Any Helm chart located in a repository.
            </Text>
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/helm-chart-components"
              target="_blank"
            >
              Learn more
            </Link>
          </Card>
          <Card>
            <Text className="!font-medium">Terraform component</Text>
            <Text variant="caption">Any Terraform module.</Text>
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/terraform-components"
              target="_blank"
            >
              Learn more
            </Link>
          </Card>
          <Card>
            <Text className="!font-medium">Jobs component</Text>
            <Text variant="caption">
              Any script that can be executed using an OCI image.
            </Text>
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/job-components"
              target="_blank"
            >
              Learn more
            </Link>
          </Card>
        </div>
      </div>
    </div>
  )
}

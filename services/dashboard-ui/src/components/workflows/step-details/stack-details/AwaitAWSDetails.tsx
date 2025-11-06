"use client";

import { Card } from "@/components/common/Card";
import { ClickToCopyButton } from "@/components/common/ClickToCopy";
import { Code } from "@/components/common/Code";
import { Divider } from "@/components/common/Divider";
import { Link } from "@/components/common/Link";
import { Skeleton } from "@/components/common/Skeleton";
import { Text } from "@/components/common/Text";
import type { IStackDetails } from "./types";

export const AwaitAWSDetails = ({ stack }: IStackDetails) => {
  return (
    <>
      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>Install quick link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.quick_link_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.quick_link_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.quick_link_url}</Code>
          </Link>
        </Card>

        <Card>
          <span className="flex justify-between items-center">
            <Text weight="strong">Install template link</Text>
            <ClickToCopyButton
              textToCopy={stack?.versions?.at(0)?.template_url}
            />
          </span>
          <Link
            href={stack?.versions?.at(0)?.template_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Code>{stack?.versions?.at(0)?.template_url}</Code>
          </Link>
        </Card>
      </div>

      <Divider dividerWord="or" />

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Setup your install stack using CLI command
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>AWS CloudFormation create stack</Text>

            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={` aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
            />
          </span>
          <Code>
            aws cloudformation create-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </Card>
      </div>

      <div className="flex flex-col gap-4">
        <Text variant="base" weight="strong">
          Update an existing install stack using CLI command
        </Text>
        <Card>
          <span className="flex justify-between items-center">
            <Text>AWS CloudFormation update stack</Text>

            <ClickToCopyButton
              className="w-fit self-end"
              textToCopy={` aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url ${stack?.versions?.at(0)?.template_url}`}
            />
          </span>
          <Code>
            aws cloudformation update-stack --stack-name [YOUR_STACK_NAME]
            --template-url {stack?.versions?.at(0)?.template_url}
          </Code>
        </Card>
      </div>
    </>
  );
};

export const AwaitAWSDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="175px" />

      <Card>
        <Skeleton height="17px" width="100px" />
        <Skeleton height="132px" width="100%" />
      </Card>

      <Card>
        <Skeleton height="17px" width="120px" />
        <Skeleton height="72px" width="100%" />
      </Card>

      <Divider dividerWord="or" />

      <Skeleton height="24px" width="325px" />

      <Card>
        <Skeleton height="17px" width="219px" />
        <Skeleton height="92px" width="100%" />
      </Card>

      <Skeleton height="24px" width="382px" />

      <Card>
        <Skeleton height="17px" width="223px" />
        <Skeleton height="92px" width="100%" />
      </Card>
    </>
  );
};

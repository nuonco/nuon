"use client";

import { Plan } from "@/components/approvals/Plan";
import { Icon } from "@/components/common/Icon";
import { Link } from "@/components/common/Link";
import { Text } from "@/components/common/Text";
import { useOrg } from "@/hooks/use-org";
import type { TWorkflowStep } from "@/types";
import { SandboxRunApply } from "./SandboxRunApply";

interface ISandboxRunStepDetails {
  step?: TWorkflowStep;
}

export const SandboxRunStepDetails = ({ step }: ISandboxRunStepDetails) => {
  const { org } = useOrg();

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Sandox run
        </Text>

        <Text variant="subtext">
          <Link
            href={`/${org.id}/installs/${step.owner_id}/sandbox`}
          >
            View sandbox <Icon variant="CaretRight" />
          </Link>
        </Text>

        <Text variant="subtext">
          <Link
            href={`/${org.id}/installs/${step.owner_id}/sandbox/${step.step_target_id}`}
          >
            View run <Icon variant="CaretRight" />
          </Link>
        </Text>
      </div>

      {step?.execution_type === "approval" ? (
        <Plan step={step} />
      ) : (
        <SandboxRunApply step={step} />
      )}
    </div>
  );
};

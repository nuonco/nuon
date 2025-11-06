"use client";

import type { TButtonSize } from "@/components/common/Button";
import { ApprovePlanButton } from "@/components/approvals/ApprovePlan";
import { DenyPlanButton } from "@/components/approvals/DenyPlan";
import type { TWorkflowStep } from "@/types";
import { getStepButtons } from "@/utils/workflow-utils";
import { RetryStepButton } from "./RetryStep";
import { SkipStepButton } from "./SkipStep";

// TODO(nnnnat): supports step cancel button but has been removed until step cancelling is added to api
export const StepButtons = ({
  buttonSize = "sm",
  isApproveAll = false,
  step,
}: {
  buttonSize?: TButtonSize;
  isApproveAll?: boolean;
  step: TWorkflowStep;
}) => {
  const { approval, retry } = getStepButtons(step);
  return (
    <div className="md:ml-auto flex items-center gap-4">
      {retry ? (
        <>
          {step?.skippable ? (
            <SkipStepButton size={buttonSize} variant="danger" step={step} />
          ) : null}
          <RetryStepButton size={buttonSize} variant="primary" step={step} />
        </>
      ) : null}
      {approval && !isApproveAll ? (
        <>
          <DenyPlanButton size={buttonSize} step={step} />
          <ApprovePlanButton size={buttonSize} variant="primary" step={step} />
        </>
      ) : null}
    </div>
  );
};

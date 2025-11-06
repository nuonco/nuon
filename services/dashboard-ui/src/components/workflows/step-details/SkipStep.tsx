"use client";

import { usePathname } from "next/navigation";
import { retryWorkflowStep } from "@/actions/workflows/retry-workflow-step";
import { Banner } from "@/components/common/Banner";
import { Button, type IButtonAsButton } from "@/components/common/Button";
import { Icon } from "@/components/common/Icon";
import { Text } from "@/components/common/Text";
import { Modal, type IModal } from "@/components/surfaces/Modal";
import { useInstall } from "@/hooks/use-install";
import { useOrg } from "@/hooks/use-org";
import { useRemovePanelByKey } from "@/hooks/use-remove-panel-by-key";
import { useSurfaces } from "@/hooks/use-surfaces";
import { useServerAction } from "@/hooks/use-server-action";
import { useServerActionToast } from "@/hooks/use-server-action-toast";
import type { TWorkflowStep } from "@/types";
import { toSentenceCase } from "@/utils/string-utils";

interface ISkipStep {
  step: TWorkflowStep;
}

export const SkipStepModal = ({ step, ...props }: ISkipStep & IModal) => {
  const path = usePathname();
  const { org } = useOrg();
  const { install } = useInstall();
  const { removeModal } = useSurfaces();
  const removePanelByKey = useRemovePanelByKey();
  const { data, error, isLoading, execute } = useServerAction({
    action: retryWorkflowStep,
  });

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>There was an error while skip this step.</Text>
        <Text>{error?.error || "Unknow error occurred."}</Text>
      </>
    ),
    errorHeading: `Failed to skip step`,
    onSuccess: () => {
      removePanelByKey(step.id);
      removeModal(props.modalId);
    },
    successContent: (
      <Text>
        {toSentenceCase(step.name)} was skipped. The workflow will continue with
        the remaining steps.
      </Text>
    ),
    successHeading: `Step skipped`,
  });

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="stronger"
        >
          Skip step?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Skipping step
          </span>
        ) : (
          "Skip step"
        ),
        onClick: () => {
          execute({
            body: {
              operation: "skip-step",
              step_id: step.id,
            },
            workflowId: step?.install_workflow_id,
            orgId: org.id,
            path,
          });
        },

        variant: "primary",
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              "An error happned, please refresh the page and try again."}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to skip this step?
        </Text>
        <Text variant="base">
          Skipping will bypass this step and continue the workflow with the
          remaining steps. Any actions or changes from this step will not be
          applied.
        </Text>
      </div>
    </Modal>
  );
};

export const SkipStepButton = ({
  step,
  ...props
}: ISkipStep & IButtonAsButton) => {
  const { addModal } = useSurfaces();
  const modal = <SkipStepModal step={step} />;

  return (
    <Button
      onClick={() => {
        addModal(modal);
      }}
      {...props}
    >
      Skip step
    </Button>
  );
};

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

interface IRetryStep {
  step: TWorkflowStep;
}

export const RetryStepModal = ({ step, ...props }: IRetryStep & IModal) => {
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
        <Text>There was an error while retrying this step.</Text>
        <Text>{error?.error || "Unknow error occurred."}</Text>
      </>
    ),
    errorHeading: `Failed to retry step`,
    onSuccess: () => {
      removePanelByKey(step.id);
      removeModal(props.modalId);
    },
    successContent: <Text>{toSentenceCase(step.name)} is being retried.</Text>,
    successHeading: `Step retry initiated`,
  });

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="stronger"
        >
          Retry step?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Retrying step
          </span>
        ) : (
          "Retry step"
        ),
        onClick: () => {
          execute({
            body: {
              operation: "retry-step",
              step_id: step.id,
            },
            workflowId: step?.workflow_id,
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
          Are you sure you want to retry this step?
        </Text>
        <Text variant="base">
          Retrying will rerun this workflow step. If successful, the workflow
          will continue from this point.
        </Text>
      </div>
    </Modal>
  );
};

export const RetryStepButton = ({
  step,
  ...props
}: IRetryStep & IButtonAsButton) => {
  const { addModal } = useSurfaces();
  const modal = <RetryStepModal step={step} />;

  return (
    <Button
      onClick={() => {
        addModal(modal);
      }}
      {...props}
    >
      Retry step
    </Button>
  );
};

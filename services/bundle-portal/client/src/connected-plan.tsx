import { Plan as PlanComponent } from "@/components/approvals/Plan/Plan";
import { CodeBlock } from "@/components/common/CodeBlock";
import { EmptyState } from "@/components/common/EmptyState";
import type { TWorkflowStep } from "@/types/ctl-api.types";

const dashboardPlanTypes = new Set([
  "terraform_plan",
  "helm_approval",
  "kubernetes_manifest_approval",
  "pulumi_plan",
]);

export const ConnectedApprovalPlan = ({
  step,
  plan,
}: {
  step: TWorkflowStep;
  plan: unknown;
}) => {
  const type = step.approval?.type;
  if (plan == null)
    return (
      <EmptyState
        variant="history"
        emptyTitle="Plan unavailable"
        emptyMessage="This step has not produced a plan yet."
      />
    );
  if (type && dashboardPlanTypes.has(type)) {
    return (
      <PlanComponent
        step={step}
        plan={plan}
        isLoading={false}
        error={undefined}
      />
    );
  }
  if (type === "noop")
    return (
      <EmptyState
        variant="history"
        emptyTitle="No changes"
        emptyMessage="No changes require review."
      />
    );
  return (
    <div className="pt-4">
      <CodeBlock language="json">{JSON.stringify(plan, null, 2)}</CodeBlock>
    </div>
  );
};

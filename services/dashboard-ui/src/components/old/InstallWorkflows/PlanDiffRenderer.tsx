'use client'

import React from 'react';
import { HelmChangesViewer } from './HelmPlanDiff';
import { K8SPlanDiff } from './K8SPlanDiff';
import { TerraformPlanViewer } from './TerraformPlanDiff';

interface PlanDiffRendererProps {
  planData: any;
  approvalType?: string;
  showNoops?: boolean;
}

/**
 * Determine if plan data has k8s diffs
 */
const hasK8SDiffs = (planData: any) => {
  return planData && 
         planData.k8s_content_diff && 
         Array.isArray(planData.k8s_content_diff) && 
         planData.k8s_content_diff.length > 0;
};

/**
 * Determine if plan data has helm diffs
 */
const hasHelmDiffs = (planData: any) => {
  return planData && 
         planData.helm_content_diff && 
         Array.isArray(planData.helm_content_diff) && 
         planData.helm_content_diff.length > 0;
};

/**
 * A component that selects the appropriate diff renderer based on the plan data
 */
export const PlanDiffRenderer: React.FC<PlanDiffRendererProps> = ({ 
  planData, 
  approvalType,
  showNoops = false 
}) => {
  // Check for k8s diffs first - they have priority
  if (hasK8SDiffs(planData) || approvalType === 'kubernetes_manifest_approval') {
    return <K8SPlanDiff planData={planData} />;
  }
  
  // Then check for helm diffs
  if (hasHelmDiffs(planData) || approvalType === 'helm_approval') {
    return <HelmChangesViewer planData={planData} />;
  }
  
  // Default to terraform plan viewer
  return <TerraformPlanViewer plan={planData} showNoops={showNoops} />;
};

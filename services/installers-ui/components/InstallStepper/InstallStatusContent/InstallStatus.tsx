"use client";

import React, { type FC } from "react";

import { Card, StatusIcon } from "@/components";

export const InstallStatus: FC<{ install: Record<string, any> }> = ({
  install,
}) => {
  let sandboxStatus = "pending";
  let sandboxStatusDescription = "waiting to provision";
  if (install.install_sandbox_runs && install.install_sandbox_runs.length > 0) {
    const lastRun = install.install_sandbox_runs[0];
    sandboxStatus = lastRun.status;
    sandboxStatusDescription = lastRun.status_description;
  }

  const install_components = install.install_components || [];
  const components = install_components.map((component, idx) => {
    let status = "pending";
    let status_description = "waiting to deploy";
    if (component.install_deploys.length > 0) {
      const lastDeploy = component.install_deploys[0];
      status = lastDeploy.status;
      status_description = lastDeploy.status_description;
    }
    return (
      <div key={idx}>
        <span className="font-medium">
          <StatusIcon status={status} /> {component.component.name}:
        </span>{" "}
        <span>{status_description}</span>
      </div>
    );
  });

  return (
    <div>
      <Card className="grid p-4 divide-y mb-4">
        <div className="font-bold">
          <StatusIcon status={sandboxStatus} /> Sandbox
        </div>
        <div>{sandboxStatusDescription}</div>
      </Card>
      <Card className="grid p-4 divide-y">
        <div className="font-bold">Components</div>
        {components}
      </Card>
    </div>
  );
};

import type {
  TAppReleaseWithFiles,
  TOTELLog,
  TWorkflow,
  TWorkflowStep,
} from "@/types/ctl-api.types";

export type TPortalBranding = {
  name: string;
  logo_url?: string;
  favicon_url?: string;
  primary_color: string;
  support_url?: string;
};

export type TConnectedRelease = {
  release: TAppReleaseWithFiles;
  active: boolean;
};

export type TConnectedReleaseUpdate = {
  id: string;
  app_release_id?: string;
  workflow_id?: string;
  created_at: string;
  status?: {
    status?: string;
    status_human_description?: string;
    metadata?: Record<string, unknown>;
  };
  workflow?: TConnectedWorkflow;
};

export type TConnectedApproval = Omit<
  NonNullable<TWorkflowStep["approval"]>,
  "id" | "type" | "response"
> & {
  id: string;
  type: NonNullable<NonNullable<TWorkflowStep["approval"]>["type"]>;
  response?: { type: string; note?: string };
};

export type TConnectedWorkflowStep = Omit<
  TWorkflowStep,
  "id" | "name" | "status" | "approval" | "log_stream"
> & {
  id: string;
  name: string;
  status?: TWorkflowStep["status"];
  approval?: TConnectedApproval;
  log_stream?: NonNullable<TWorkflowStep["log_stream"]> & { id: string };
};

export type TConnectedLog = TOTELLog & {
  id: string;
  timestamp: string;
  body: string;
};

export type TConnectedWorkflow = Omit<
  TWorkflow,
  "id" | "type" | "created_at" | "status" | "steps"
> & {
  id: string;
  name?: string;
  type: NonNullable<TWorkflow["type"]>;
  created_at: string;
  started_at?: string;
  finished_at?: string;
  status?: TWorkflow["status"];
  steps?: TConnectedWorkflowStep[];
};

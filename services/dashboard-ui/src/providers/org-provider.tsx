"use client";

import { createContext, useEffect, type ReactNode } from "react";
import { setOrgCookie } from "@/actions/orgs/org-session-cookie";
import { usePolling, type IPollingProps } from "@/hooks/use-polling";
import type { TOrg } from "@/types";

type OrgContextValue = {
  org: TOrg | null;
  isLoading: boolean;
  error: any;
  refresh: () => void;
};

export const OrgContext = createContext<OrgContextValue | undefined>(undefined);

export function OrgProvider({
  children,
  initOrg,
  pollInterval = 30000,
  shouldPoll = false,
}: {
  children: ReactNode;
  initOrg: TOrg;
} & IPollingProps) {
  const {
    data: org,
    error,
    isLoading,
  } = usePolling<TOrg>({
    initData: initOrg,
    path: `/api/orgs/${initOrg.id}`,
    pollInterval,
    shouldPoll,
  });

  useEffect(() => {
    setOrgCookie(initOrg?.id);
  }, [initOrg?.id]);

  return (
    <OrgContext.Provider
      value={{
        org,
        isLoading,
        error,
        refresh: () => {
          /* implement if needed */
        },
      }}
    >
      {children}
    </OrgContext.Provider>
  );
}

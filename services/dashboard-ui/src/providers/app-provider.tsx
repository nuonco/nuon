"use client";

import { createContext, type ReactNode } from "react";
import { usePolling, type IPollingProps } from "@/hooks/use-polling";
import { useOrg } from "@/hooks/use-org";
import type { TApp } from "@/types";

type AppContextValue = {
  app: TApp | null;
  isLoading: boolean;
  error: any;
  refresh: () => void;
};

export const AppContext = createContext<AppContextValue | undefined>(undefined);

export function AppProvider({
  children,
  initApp,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode;
  initApp: TApp;
} & IPollingProps) {
  const { org } = useOrg();
  const {
    data: app,
    error,
    isLoading,
  } = usePolling<TApp>({
    initData: initApp,
    path: `/api/orgs/${org.id}/apps/${initApp.id}`,
    pollInterval,
    shouldPoll,
  });

  return (
    <AppContext.Provider
      value={{
        app,
        isLoading,
        error,
        refresh: () => {
          /* implement if needed */
        },
      }}
    >
      {children}
    </AppContext.Provider>
  );
}

import "./fonts.css";
import "../src/app/globals.css";
import { GlobalProvider } from "@ladle/react";
import { AppRouterContext } from "next/dist/shared/lib/app-router-context.shared-runtime";
import { MockNavigationProvider } from "./NextNavigationMocks";

export const Provider: GlobalProvider = ({ children }) => {
  return (
    <AppRouterContext.Provider
      value={{
        back: () => {
          console.log("Router: back called");
        },
        forward: () => {
          console.log("Router: forward called");
        },
        prefetch: () => {
          console.log("Router: prefetch called");
        },
        push: (href: string) => {
          console.log("Router: push called with:", href);
        },
        refresh: () => {
          console.log("Router: refresh called");
        },
        replace: (href: string) => {
          console.log("Router: replace called with:", href);
        },
      }}
    >
      <MockNavigationProvider>
        {children}
      </MockNavigationProvider>
    </AppRouterContext.Provider>
  );
};

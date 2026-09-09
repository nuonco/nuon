import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter } from "react-router";
import { SurfacesProvider } from "@/providers/surfaces-provider";
import { BrandingBoundary } from "./branding";
import { ConnectedApp } from "./ConnectedApp";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <SurfacesProvider>
          <BrandingBoundary>
            <ConnectedApp />
          </BrandingBoundary>
        </SurfacesProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);

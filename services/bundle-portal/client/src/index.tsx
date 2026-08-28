import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter } from "react-router";
import { SurfacesProvider } from "@/providers/surfaces-provider";
import { App } from "./App";
import { BrandingBoundary } from "./branding";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchInterval: 5000,
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
            <App />
          </BrandingBoundary>
        </SurfacesProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);

// Comprehensive Next.js navigation mocks for Ladle
import React from "react";

// Mock data that can be customized
const MOCK_PATHNAME = "/mock-org/mock-path";
const MOCK_SEARCH_PARAMS = new URLSearchParams("?tab=overview&status=active");
const MOCK_PARAMS = {
  orgId: "mock-org-id",
  installId: "mock-install-id",
  appId: "mock-app-id",
  configId: "mock-config-id",
  workflowId: "mock-workflow-id",
  runId: "mock-run-id",
  deployId: "mock-deploy-id",
  componentId: "mock-component-id",
  buildId: "mock-build-id",
  actionId: "mock-action-id",
};

// useParams mock
export const useParams = () => {
  console.log("Mock useParams called, returning:", MOCK_PARAMS);
  return MOCK_PARAMS;
};

// useRouter mock
export const useRouter = () => {
  const router = {
    push: (href: string, options?: any) => {
      console.log("Mock router.push called with:", { href, options });
    },
    replace: (href: string, options?: any) => {
      console.log("Mock router.replace called with:", { href, options });
    },
    refresh: () => {
      console.log("Mock router.refresh called");
    },
    back: () => {
      console.log("Mock router.back called");
    },
    forward: () => {
      console.log("Mock router.forward called");
    },
    prefetch: (href: string) => {
      console.log("Mock router.prefetch called with:", href);
      return Promise.resolve();
    },
    // Additional router properties that might be accessed
    pathname: MOCK_PATHNAME,
    query: Object.fromEntries(MOCK_SEARCH_PARAMS),
    asPath: `${MOCK_PATHNAME}?${MOCK_SEARCH_PARAMS.toString()}`,
    route: MOCK_PATHNAME,
    isReady: true,
    isPreview: false,
  };
  
  console.log("Mock useRouter called, returning router object");
  return router;
};

// useSearchParams mock
export const useSearchParams = () => {
  const searchParams = {
    get: (key: string) => {
      const value = MOCK_SEARCH_PARAMS.get(key);
      console.log(`Mock searchParams.get("${key}") called, returning:`, value);
      return value;
    },
    getAll: (key: string) => {
      const values = MOCK_SEARCH_PARAMS.getAll(key);
      console.log(`Mock searchParams.getAll("${key}") called, returning:`, values);
      return values;
    },
    has: (key: string) => {
      const has = MOCK_SEARCH_PARAMS.has(key);
      console.log(`Mock searchParams.has("${key}") called, returning:`, has);
      return has;
    },
    keys: () => MOCK_SEARCH_PARAMS.keys(),
    values: () => MOCK_SEARCH_PARAMS.values(),
    entries: () => MOCK_SEARCH_PARAMS.entries(),
    forEach: (callback: any) => MOCK_SEARCH_PARAMS.forEach(callback),
    toString: () => MOCK_SEARCH_PARAMS.toString(),
    [Symbol.iterator]: () => MOCK_SEARCH_PARAMS[Symbol.iterator](),
  };
  
  console.log("Mock useSearchParams called, returning searchParams object");
  return searchParams;
};

// usePathname mock
export const usePathname = () => {
  console.log("Mock usePathname called, returning:", MOCK_PATHNAME);
  return MOCK_PATHNAME;
};

// Additional Next.js navigation functions that might be used
export const redirect = (url: string) => {
  console.log("Mock redirect called with:", url);
  throw new Error(`Mock redirect to: ${url}`);
};

export const notFound = () => {
  console.log("Mock notFound called");
  throw new Error("Mock not found");
};

export const permanentRedirect = (url: string) => {
  console.log("Mock permanentRedirect called with:", url);
  throw new Error(`Mock permanent redirect to: ${url}`);
};

// Hook to customize mock data in stories if needed
export const useMockNavigation = () => {
  return {
    setMockPathname: (pathname: string) => {
      console.log("Setting mock pathname to:", pathname);
      // In a real implementation, this would update the mock state
    },
    setMockParams: (params: Record<string, string>) => {
      console.log("Setting mock params to:", params);
      // In a real implementation, this would update the mock state
    },
    setMockSearchParams: (searchParams: string | URLSearchParams) => {
      console.log("Setting mock search params to:", searchParams);
      // In a real implementation, this would update the mock state
    },
  };
};

// Provider component to wrap stories that need navigation context
export const MockNavigationProvider = ({ children }: { children: React.ReactNode }) => {
  return (
    <div data-mock-navigation="true">
      {children}
    </div>
  );
};
"use client";

import { useCallback, useState } from "react";
import type { TAPIError, TAPIResponse } from "@/types";

interface IUseServerAction<TArgs extends any[], TData> {
  action: (...args: TArgs) => Promise<TAPIResponse<TData>>;
}

export function useServerAction<TArgs extends any[], TData>({
  action,
}: IUseServerAction<TArgs, TData>) {
  const [data, setData] = useState<TData | null>(null);
  const [error, setError] = useState<TAPIError | null>(null);
  const [headers, setHeaders] = useState<Headers | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [status, setStatus] = useState<number | null>(null);

  // Call this to invoke your server action
  const execute = useCallback(
    async (...args: TArgs): Promise<TAPIResponse<TData>> => {
      setIsLoading(true);
      setError(null);
      setStatus(null);
      setHeaders(null);

      try {
        const response = await action(...args);

        // Convert array of headers to Headers object if necessary
        let hdrs: Headers | null = null;
        if (Array.isArray(response.headers)) {
          hdrs = new Headers(response.headers as [string, string][]);
        } else if (response.headers instanceof Headers) {
          hdrs = response.headers;
        }
        
        setData(response.data);
        setError(response.error);
        setStatus(response.status);
        setHeaders(hdrs);
        return response;
      } catch (err: any) {
        setData(null);
        setError(err);
        setStatus(null);
        setHeaders(null);

        // Return a generic error response if the action throws
        return {
          data: null,
          error: err,
          status: null as any,
          headers: null as any,
        };
      } finally {
        setIsLoading(false);
      }
    },
    [action]
  );

  return { data, error, status, headers, isLoading, execute };
}

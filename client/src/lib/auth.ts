import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';

export interface FetchOptions extends RequestInit {
  headers?: Record<string, string>;
}

// Track if we're currently refreshing to prevent multiple refreshes
let isRefreshing = false;

// Queue of requests waiting for token refresh
let refreshQueue: Array<() => void> = [];

// Process queued requests after token refresh
function processQueue() {
  refreshQueue.forEach((callback) => callback());
  refreshQueue = [];
}

// Combined hook that handles authentication, token refresh and CSRF
export const useAuthCheck = () => {
  const [csrfToken, setCsrfToken] = useState<string>('');
  const [csrfLoading, setCsrfLoading] = useState<boolean>(true);
  const [csrfError, setCsrfError] = useState<Error | null>(null);
  const queryClient = useQueryClient();

  // Fetch CSRF token
  useEffect(() => {
    async function fetchCsrfToken() {
      try {
        setCsrfLoading(true);
        const response = await fetch('/api/csrf-token', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error(
            `Failed to fetch CSRF token: ${response.status}`
          );
        }

        const data = await response.json();
        setCsrfToken(data.csrf_token);
        setCsrfError(null);
      } catch (err) {
        console.error('Error fetching CSRF token:', err);
        setCsrfError(
          err instanceof Error ? err : new Error(String(err))
        );
      } finally {
        setCsrfLoading(false);
      }
    }

    fetchCsrfToken();
  }, []);

  // Function to refresh access token
  const refreshAccessToken = async (): Promise<boolean> => {
    try {
      const response = await fetch('/api/refresh-token', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'X-CSRF-Token': csrfToken,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to refresh access token');
      }

      return true;
    } catch (err) {
      console.error('Token refresh failed:', err);
      return false;
    }
  };

  // Fetch function that includes CSRF token and handles token refresh
  const fetchWithCsrf = async (
    url: string,
    options: FetchOptions = {}
  ) => {
    try {
      // First attempt with current tokens
      const headers = {
        ...options.headers,
        'X-CSRF-Token': csrfToken,
      };

      const response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include',
      });

      // If not unauthorized, return response
      if (response.status !== 401) {
        return response;
      }

      // Check if it's a token expiration error
      try {
        const data = await response.json();
        if (data.code !== 'token_expired') {
          // Not a token expiration issue
          return response;
        }
      } catch {
        // If we can't parse JSON, it's not a token expiration error
        return response;
      }

      // Handle token refresh
      let refreshSuccess = false;

      // If not already refreshing, initiate refresh
      if (!isRefreshing) {
        isRefreshing = true;
        refreshSuccess = await refreshAccessToken();
        isRefreshing = false;
        processQueue();
      } else {
        // Wait for the ongoing refresh to complete
        await new Promise<void>((resolve) => {
          refreshQueue.push(resolve);
        });
      }

      // Retry original request with new token
      return fetch(url, {
        ...options,
        headers,
        credentials: 'include',
      });
    } catch (error) {
      console.error('Request failed:', error);
      throw error;
    }
  };

  // Use the fetchWithCsrf in auth check
  const {
    data: user,
    isLoading: authLoading,
    error: authError,
  } = useQuery({
    queryKey: ['authUser'],
    queryFn: async () => {
      // Skip the check if CSRF token is still loading
      if (csrfLoading) {
        return null;
      }

      const response = await fetchWithCsrf('/api/profile');

      if (!response.ok) {
        if (response.status === 401) {
          return null; // Not authenticated
        }
        throw new Error('Failed to fetch user');
      }

      return response.json();
    },
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    // Only run the query once we have a CSRF token
    enabled: !csrfLoading,
  });

  // Function to refresh the CSRF token
  const refreshCsrfToken = async () => {
    try {
      setCsrfLoading(true);
      const response = await fetch('/api/csrf-token', {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error(
          `Failed to refresh CSRF token: ${response.status}`
        );
      }

      const data = await response.json();
      setCsrfToken(data.csrf_token);
      setCsrfError(null);
      return data.csrf_token;
    } catch (err) {
      console.error('Error refreshing CSRF token:', err);
      setCsrfError(
        err instanceof Error ? err : new Error(String(err))
      );
      throw err;
    } finally {
      setCsrfLoading(false);
    }
  };

  // Function to refresh the user data
  const refreshUser = () => {
    queryClient.invalidateQueries({ queryKey: ['authUser'] });
  };

  // Logout function
  const logout = async () => {
    try {
      await fetchWithCsrf('/api/logout', { method: 'POST' });
      queryClient.setQueryData(['authUser'], null);
    } catch (error) {
      console.error('Logout error:', error);
      throw error;
    }
  };

  return {
    user,
    isLoading: authLoading || csrfLoading,
    error: authError || csrfError,
    csrfToken,
    logout,
    refreshUser,
    refreshCsrfToken,
    fetchWithCsrf,
  };
};

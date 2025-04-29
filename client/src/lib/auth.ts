import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';

export interface FetchOptions extends RequestInit {
  headers?: Record<string, string>;
}

// Track if we're currently refreshing to prevent multiple refreshes
let isRefreshing = false;

// Queue of requests waiting for token refresh
let refreshQueue: Array<(err?: Error | null) => void> = [];

// Combined hook that handles authentication, token refresh and CSRF
export const useAuthCheck = () => {
  const [csrfToken, setCsrfToken] = useState('');
  const [csrfLoading, setCsrfLoading] = useState(true);
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
    // Add CSRF token to headers
    const headers = {
      ...options.headers,
      'X-CSRF-Token': csrfToken,
    };

    // First attempt
    const response = await fetch(url, {
      ...options,
      headers,
      credentials: 'include',
    });

    // Check if it's a token expiration error
    if (response.status === 401) {
      try {
        const data = await response.json();
        if (data.code === 'token_expired') {
          // Handle token refresh
          if (!isRefreshing) {
            isRefreshing = true;
            try {
              const ok = await refreshAccessToken();
              if (!ok) throw new Error('Token refresh failed');
              // Wake up all waiters successfully
              refreshQueue.forEach((cb) => cb(null));
            } catch (err) {
              const typedError =
                err instanceof Error ? err : new Error(String(err));
              // Wake up all waiters with error
              refreshQueue.forEach((cb) => cb(typedError));
              throw err;
            } finally {
              isRefreshing = false;
              refreshQueue = [];
            }
          } else {
            // Wait for the in-flight refresh to resolve or reject
            await new Promise<void>((resolve, reject) => {
              refreshQueue.push((err?) =>
                err ? reject(err) : resolve()
              );
            });
          }

          // Retry original request with new token
          return fetch(url, {
            ...options,
            headers,
            credentials: 'include',
          });
        }
      } catch {
        // Ignore errors parsing the response body
      }
    }

    return response;
  };

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

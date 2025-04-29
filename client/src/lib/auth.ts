import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';

export interface FetchOptions extends RequestInit {
  headers?: Record<string, string>;
}

// Combined hook that handles both authentication and CSRF token
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

  // Fetch function that includes the CSRF token
  const fetchWithCsrf = (url: string, options: FetchOptions = {}) => {
    const headers = {
      ...options.headers,
      'X-CSRF-Token': csrfToken,
    };

    return fetch(url, {
      ...options,
      headers,
      credentials: 'include',
    });
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

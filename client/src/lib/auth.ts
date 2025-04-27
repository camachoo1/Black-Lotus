// lib/auth.ts
import { useQuery } from '@tanstack/react-query';

export const useAuthCheck = () => {
  return useQuery({
    queryKey: ['authUser'],
    queryFn: async () => {
      const response = await fetch('/api/profile', {
        credentials: 'include',
      });

      if (!response.ok) {
        if (response.status === 401) {
          return null; // Not authenticated
        }
        throw new Error('Failed to fetch user');
      }

      return response.json();
    },
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });
};
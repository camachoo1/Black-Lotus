import { createContext, useContext, ReactNode } from 'react';
import {
  useQueryClient,
  useMutation,
} from '@tanstack/react-query';
import { useAuthCheck } from '@/lib/auth';

interface User {
  id: string;
  name: string;
  email: string;
  email_verified: boolean;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  logout: () => Promise<void>;
  refreshUser: () => void;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  isLoading: true,
  logout: async () => {},
  refreshUser: () => {},
});

export const AuthProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const queryClient = useQueryClient();
  const { data: user, isLoading } = useAuthCheck();

  // Logout mutation
  const logoutMutation = useMutation({
    mutationFn: async () => {
      await fetch('/api/logout', {
        method: 'POST',
        credentials: 'include',
      });
    },
    onSuccess: () => {
      queryClient.setQueryData(['authUser'], null);
    },
  });

  const refreshUser = () => {
    queryClient.invalidateQueries({ queryKey: ['authUser'] });
  };

  const logout = async () => {
    await logoutMutation.mutateAsync();
  };

  return (
    <AuthContext.Provider
      value={{
        user: user ?? null,
        isLoading,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);

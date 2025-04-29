import { FetchOptions, useAuthCheck } from '@/lib/auth';
import { createContext, useContext, ReactNode } from 'react';

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
  fetchWithCsrf: (url: string, options?: FetchOptions) => Promise<Response>;
  csrfToken: string;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  isLoading: true,
  logout: async () => {},
  refreshUser: () => {},
  fetchWithCsrf: async () => new Response(),
  csrfToken: '',
});

export const AuthProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const {
    user,
    isLoading,
    logout: logoutFromHook,
    refreshUser: refreshUserFromHook,
    fetchWithCsrf,
    csrfToken,
  } = useAuthCheck();


  const refreshUser = () => {
    refreshUserFromHook();
  };

  const logout = async () => {
    await logoutFromHook();
  };

  return (
    <AuthContext.Provider
      value={{
        user: user ?? null,
        isLoading,
        logout,
        refreshUser,
        fetchWithCsrf,
        csrfToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);

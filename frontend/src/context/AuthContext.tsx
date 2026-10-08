import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { apiClient, storeTokens, purgeStoredTokens, getStoredAccessToken, onAuthExpired } from '../api/client';
import { User, LoginResponseData, ApiResponse } from '../types/auth';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const logout = useCallback(() => {
    purgeStoredTokens();
    setUser(null);
  }, []);

  const refreshUser = useCallback(async () => {
    try {
      const response = await apiClient.get<ApiResponse<User>>('/auth/me');
      if (response.data && response.data.success && response.data.data) {
        setUser(response.data.data);
      } else {
        logout();
      }
    } catch {
      logout();
    }
  }, [logout]);

  const login = async (credentials: LoginCredentials): Promise<void> => {
    const response = await apiClient.post<ApiResponse<LoginResponseData>>('/auth/login', credentials);
    if (response.data && response.data.success && response.data.data) {
      const { user: loggedUser, tokens } = response.data.data;
      storeTokens(tokens);
      setUser(loggedUser);
    } else {
      throw new Error(response.data?.error?.message || 'Login fallito');
    }
  };

  useEffect(() => {
    // Sottoscrizione al segnale di scadenza sessione da parte dell'interceptor Axios
    const unsubscribe = onAuthExpired(() => {
      logout();
    });

    // Inizializzazione sessione al caricamento dell'applicazione
    const initAuth = async () => {
      const token = getStoredAccessToken();
      if (token) {
        await refreshUser();
      }
      setIsLoading(false);
    };

    initAuth();

    return () => {
      unsubscribe();
    };
  }, [refreshUser, logout]);

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth deve essere utilizzato all\'interno di un AuthProvider');
  }
  return context;
};

import React, { createContext, useContext, useState, useEffect } from 'react';

interface User {
  signer_uuid: string;
  public_key: string;
  status: string;
  signer_approval_url?: string;
  fid?: number;
}

interface UserContextType {
  user: User | null;
  login: () => Promise<void>;
  logout: () => void;
  checkSignerStatus: () => Promise<void>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const storedUser = localStorage.getItem('farcasterUser');
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
  }, []);

  const login = async () => {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/login`, {
        method: 'POST',
      });
      if (!response.ok) {
        throw new Error('Login failed');
      }
      const newUser = await response.json();
      setUser(newUser);
      localStorage.setItem('farcasterUser', JSON.stringify(newUser));
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('farcasterUser');
  };

  const checkSignerStatus = async () => {
    if (user && user.status === 'pending_approval') {
      try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/signer-status?signer_uuid=${user.signer_uuid}`);
        if (!response.ok) {
          throw new Error('Failed to check signer status');
        }
        const updatedUser = await response.json();
        if (updatedUser.status === 'approved') {
          setUser(updatedUser);
          localStorage.setItem('farcasterUser', JSON.stringify(updatedUser));
        }
      } catch (error) {
        console.error('Error checking signer status:', error);
      }
    }
  };

  return (
    <UserContext.Provider value={{ user, login, logout, checkSignerStatus }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
};
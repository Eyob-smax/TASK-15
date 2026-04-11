import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

interface OfflineStatusValue {
  isOnline: boolean;
  isOffline: boolean;
  lastSyncAt: number | null;
}

const OfflineStatusContext = createContext<OfflineStatusValue | null>(null);

export const OFFLINE_MUTATION_MESSAGE = 'Reconnect to make changes.';

interface OfflineStatusProviderProps {
  children: ReactNode;
  lastSyncAt: number | null;
}

export function OfflineStatusProvider({ children, lastSyncAt }: OfflineStatusProviderProps) {
  const [isOnline, setIsOnline] = useState(() => navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  const value = useMemo<OfflineStatusValue>(() => ({
    isOnline,
    isOffline: !isOnline,
    lastSyncAt,
  }), [isOnline, lastSyncAt]);

  return (
    <OfflineStatusContext.Provider value={value}>
      {children}
    </OfflineStatusContext.Provider>
  );
}

export function useOfflineStatus(): OfflineStatusValue {
  const context = useContext(OfflineStatusContext);
  if (!context) {
    throw new Error('useOfflineStatus must be used within an OfflineStatusProvider');
  }
  return context;
}

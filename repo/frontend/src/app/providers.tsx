import { useEffect, useRef, useState, type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import theme from '@/app/theme';
import { AuthProvider } from '@/lib/auth';
import { OfflineStatusProvider } from '@/lib/offline';
import {
  hydrateOfflineQueryCache,
  loadPersistedQuerySnapshot,
  persistOfflineQueryCache,
} from '@/lib/offline-cache';
import { replayOfflineMutations } from '@/lib/offline-mutations';
import { NotificationsProvider } from '@/lib/notifications';
import { ErrorBoundary } from '@/components/ErrorBoundary';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000,
      refetchOnWindowFocus: false,
      retry: 1,
      refetchOnReconnect: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

interface ProvidersProps {
  children: ReactNode;
}

export default function Providers({ children }: ProvidersProps) {
  const [isHydrated, setIsHydrated] = useState(false);
  const [lastSyncAt, setLastSyncAt] = useState<number | null>(null);
  const persistTimerRef = useRef<number | null>(null);

  useEffect(() => {
    let active = true;
    let unsubscribe: (() => void) | undefined;

    const schedulePersist = () => {
      if (persistTimerRef.current !== null) {
        window.clearTimeout(persistTimerRef.current);
      }
      persistTimerRef.current = window.setTimeout(async () => {
        try {
          const persistedAt = await persistOfflineQueryCache(queryClient);
          if (active && persistedAt) {
            setLastSyncAt(persistedAt);
          }
        } catch {
          // Offline persistence is best-effort and should never block app use.
        }
      }, 250);
    };

    const replayQueuedMutations = async () => {
      try {
        await replayOfflineMutations(queryClient);
      } catch {
        // Replay is best-effort and retried on the next reconnect.
      }
    };

    (async () => {
      try {
        const snapshot = await loadPersistedQuerySnapshot();
        if (active) {
          hydrateOfflineQueryCache(queryClient, snapshot);
          setLastSyncAt(snapshot?.updatedAt ?? null);
        }
      } catch {
        // If persistence cannot be read we fall back to the live API path.
      } finally {
        if (active) {
          unsubscribe = queryClient.getQueryCache().subscribe(() => {
            schedulePersist();
          });
          window.addEventListener('online', replayQueuedMutations);
          void replayQueuedMutations();
          setIsHydrated(true);
        }
      }
    })();

    return () => {
      active = false;
      unsubscribe?.();
      window.removeEventListener('online', replayQueuedMutations);
      if (persistTimerRef.current !== null) {
        window.clearTimeout(persistTimerRef.current);
      }
    };
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {isHydrated ? (
          <OfflineStatusProvider lastSyncAt={lastSyncAt}>
            <AuthProvider>
              <NotificationsProvider>
                <ErrorBoundary>
                  {children}
                </ErrorBoundary>
              </NotificationsProvider>
            </AuthProvider>
          </OfflineStatusProvider>
        ) : null}
      </ThemeProvider>
    </QueryClientProvider>
  );
}

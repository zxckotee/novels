'use client';

import { ReactNode, useEffect, useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { useAuthStore } from '@/store/auth';

interface ProvidersProps {
  children: ReactNode;
}

export function Providers({ children }: ProvidersProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // 1 минута
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      })
  );

  // Zustand persist: hydrate auth store only after mount to avoid
  // server/client HTML mismatches (hydration errors).
  useEffect(() => {
    // Rehydrate persisted auth state (localStorage) on client.
    useAuthStore.persist.rehydrate();

    // Mark auth store as loaded after hydration.
    const unsub =
      typeof useAuthStore.persist.onFinishHydration === 'function'
        ? useAuthStore.persist.onFinishHydration(() => {
            useAuthStore.getState().setLoading(false);
          })
        : undefined;

    // If already hydrated, ensure loading is false.
    if (typeof useAuthStore.persist.hasHydrated === 'function' && useAuthStore.persist.hasHydrated()) {
      useAuthStore.getState().setLoading(false);
    }

    return () => {
      if (typeof unsub === 'function') unsub();
    };
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster
        position="bottom-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#1e1e1e',
            color: '#ffffff',
            border: '1px solid #2a2a2a',
          },
          success: {
            iconTheme: {
              primary: '#22c55e',
              secondary: '#ffffff',
            },
          },
          error: {
            iconTheme: {
              primary: '#ef4444',
              secondary: '#ffffff',
            },
          },
        }}
      />
    </QueryClientProvider>
  );
}

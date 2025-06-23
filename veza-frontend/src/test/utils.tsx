import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  );
};

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options });

// Réinitialiser les mocks entre les tests
export const resetMocks = () => {
  localStorage.clear();
  sessionStorage.clear();
  queryClient.clear();
};

// Fonction utilitaire pour simuler un délai
export const wait = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

// Fonction utilitaire pour simuler une erreur API
export const mockApiError = (status: number, message: string) => {
  return {
    response: {
      status,
      data: {
        message,
      },
    },
  };
};

export * from '@testing-library/react';
export { customRender as render }; 
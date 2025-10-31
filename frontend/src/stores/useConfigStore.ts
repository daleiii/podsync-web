import { create } from 'zustand';
import { configAPI, handleAPIError } from '../services/api';
import type { AppConfig } from '../types/api';

interface ConfigStore {
  config: AppConfig | null;
  loading: boolean;
  error: string | null;
  loadConfig: () => Promise<void>;
  getBackendURL: () => string;
}

export const useConfigStore = create<ConfigStore>((set, get) => ({
  config: null,
  loading: false,
  error: null,

  loadConfig: async () => {
    set({ loading: true, error: null });
    try {
      const response = await configAPI.getConfig();
      set({ config: response.data, loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  getBackendURL: () => {
    const config = get().config;
    if (!config?.server) {
      return 'http://localhost:3000'; // Fallback
    }
    return `${config.server.hostname}:${config.server.port}`;
  },
}));

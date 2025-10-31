import { create } from 'zustand';
import { historyAPI, handleAPIError, type HistoryListParams } from '../services/api';
import type { HistoryListResponse, HistoryStatsResponse } from '../types/api';

interface HistoryStore {
  historyData: HistoryListResponse | null;
  stats: HistoryStatsResponse | null;
  loading: boolean;
  error: string | null;
  loadHistory: (params?: HistoryListParams) => Promise<void>;
  loadStats: () => Promise<void>;
  deleteHistory: (id: string) => Promise<void>;
  deleteAllHistory: () => Promise<void>;
  cleanup: () => Promise<void>;
}

export const useHistoryStore = create<HistoryStore>((set, get) => ({
  historyData: null,
  stats: null,
  loading: false,
  error: null,

  loadHistory: async (params?: HistoryListParams) => {
    set({ loading: true, error: null });
    try {
      const response = await historyAPI.listHistory(params);
      set({ historyData: response.data, loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  loadStats: async () => {
    try {
      const response = await historyAPI.getStats();
      set({ stats: response.data });
    } catch (error) {
      set({ error: handleAPIError(error) });
    }
  },

  deleteHistory: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await historyAPI.deleteHistory(id);
      // Remove the entry from the current data
      const currentData = get().historyData;
      if (currentData) {
        const updatedEntries = currentData.entries.filter((e) => e.id !== id);
        set({
          historyData: {
            ...currentData,
            entries: updatedEntries,
            total: currentData.total - 1,
          },
          loading: false,
        });
      } else {
        set({ loading: false });
      }
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
      throw error;
    }
  },

  deleteAllHistory: async () => {
    set({ loading: true, error: null });
    try {
      await historyAPI.deleteAllHistory();
      set({
        historyData: {
          entries: [],
          total: 0,
          page: 1,
          page_size: 50,
          total_pages: 0,
        },
        loading: false,
      });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
      throw error;
    }
  },

  cleanup: async () => {
    set({ loading: true, error: null });
    try {
      await historyAPI.cleanup();
      // Reload history after cleanup
      await get().loadHistory();
      set({ loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
      throw error;
    }
  },
}));

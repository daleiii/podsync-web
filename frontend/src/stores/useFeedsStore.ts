import { create } from 'zustand';
import { feedsAPI, handleAPIError } from '../services/api';
import type { Feed } from '../types/api';

interface FeedsStore {
  feeds: Feed[];
  loading: boolean;
  error: string | null;
  loadFeeds: () => Promise<void>;
  deleteFeed: (id: string) => Promise<void>;
  refreshFeed: (id: string) => Promise<void>;
}

export const useFeedsStore = create<FeedsStore>((set, get) => ({
  feeds: [],
  loading: false,
  error: null,

  loadFeeds: async () => {
    set({ loading: true, error: null });
    try {
      const response = await feedsAPI.listFeeds();
      set({ feeds: response.data, loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  deleteFeed: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await feedsAPI.deleteFeed(id);
      const updatedFeeds = get().feeds.filter((f) => f.id !== id);
      set({ feeds: updatedFeeds, loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  refreshFeed: async (id: string) => {
    try {
      await feedsAPI.refreshFeed(id);
    } catch (error) {
      set({ error: handleAPIError(error) });
      throw error;
    }
  },
}));

import { create } from 'zustand';
import { episodesAPI, handleAPIError, type EpisodeListParams } from '../services/api';
import type { EpisodeListResponse } from '../types/api';

interface EpisodesStore {
  episodesData: EpisodeListResponse | null;
  loading: boolean;
  error: string | null;
  loadEpisodes: (params?: EpisodeListParams) => Promise<void>;
  deleteEpisode: (feedId: string, episodeId: string) => Promise<void>;
  retryEpisode: (feedId: string, episodeId: string) => Promise<void>;
  blockEpisode: (feedId: string, episodeId: string) => Promise<void>;
}

export const useEpisodesStore = create<EpisodesStore>((set, get) => ({
  episodesData: null,
  loading: false,
  error: null,

  loadEpisodes: async (params?: EpisodeListParams) => {
    set({ loading: true, error: null });
    try {
      const response = await episodesAPI.listEpisodes(params);
      set({ episodesData: response.data, loading: false });
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  deleteEpisode: async (feedId: string, episodeId: string) => {
    set({ loading: true, error: null });
    try {
      await episodesAPI.deleteEpisode(feedId, episodeId);
      // Reload episodes after deletion
      const currentData = get().episodesData;
      if (currentData) {
        const updatedEpisodes = currentData.episodes.filter(
          (e) => !(e.feed_id === feedId && e.id === episodeId)
        );
        set({
          episodesData: {
            ...currentData,
            episodes: updatedEpisodes,
            total: currentData.total - 1,
          },
          loading: false,
        });
      } else {
        set({ loading: false });
      }
    } catch (error) {
      set({ error: handleAPIError(error), loading: false });
    }
  },

  retryEpisode: async (feedId: string, episodeId: string) => {
    try {
      await episodesAPI.retryEpisode(feedId, episodeId);
      // Update the episode status optimistically to 'new'
      const currentData = get().episodesData;
      if (currentData) {
        const updatedEpisodes = currentData.episodes.map((e) => {
          if (e.feed_id === feedId && e.id === episodeId) {
            return { ...e, status: 'new' as const, error: '' };
          }
          return e;
        });
        set({
          episodesData: {
            ...currentData,
            episodes: updatedEpisodes,
          },
        });
      }
    } catch (error) {
      set({ error: handleAPIError(error) });
      throw error;
    }
  },

  blockEpisode: async (feedId: string, episodeId: string) => {
    set({ loading: true, error: null });
    try {
      await episodesAPI.blockEpisode(feedId, episodeId);
      // Update the episode status to 'blocked'
      const currentData = get().episodesData;
      if (currentData) {
        const updatedEpisodes = currentData.episodes.map((e) => {
          if (e.feed_id === feedId && e.id === episodeId) {
            return { ...e, status: 'blocked' as const, error: '' };
          }
          return e;
        });
        set({
          episodesData: {
            ...currentData,
            episodes: updatedEpisodes,
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
}));

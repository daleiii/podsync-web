import axios from 'axios';
import type {
  AppConfig,
  Feed,
  EpisodeListResponse,
  HistoryEntry,
  HistoryFilters,
  HistoryListResponse,
  HistoryStatsResponse,
} from '../types/api';

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Configuration API
export const configAPI = {
  getConfig: () => api.get<AppConfig>('/config'),
};

// Feeds API
export const feedsAPI = {
  listFeeds: () => api.get<Feed[]>('/feeds'),
  getFeed: (id: string) => api.get<Feed>(`/feeds/${id}`),
  deleteFeed: (id: string) => api.delete(`/feeds/${id}`),
  refreshFeed: (id: string) => api.post(`/feeds/${id}/refresh`),
};

// Episodes API
export interface EpisodeListParams {
  page?: number;
  page_size?: number;
  feed_id?: string;
  status?: string;
  search?: string;
  show_ignored?: boolean;
  date_filter?: string;
  date_start?: string;
  date_end?: string;
}

export const episodesAPI = {
  listEpisodes: (params?: EpisodeListParams) =>
    api.get<EpisodeListResponse>('/episodes', { params }),
  deleteEpisode: (feedId: string, episodeId: string) =>
    api.delete(`/episodes/${feedId}/${episodeId}`),
  retryEpisode: (feedId: string, episodeId: string) =>
    api.post(`/episodes/${feedId}/${episodeId}/retry`),
  blockEpisode: (feedId: string, episodeId: string) =>
    api.post(`/episodes/${feedId}/${episodeId}/block`),
};

// History API
export interface HistoryListParams extends HistoryFilters {
  page?: number;
  page_size?: number;
}

export const historyAPI = {
  listHistory: (params?: HistoryListParams) =>
    api.get<HistoryListResponse>('/history', { params }),
  getHistory: (id: string) =>
    api.get<HistoryEntry>(`/history/${id}`),
  deleteHistory: (id: string) =>
    api.delete(`/history/${id}`),
  deleteAllHistory: () =>
    api.delete('/history'),
  getStats: () =>
    api.get<HistoryStatsResponse>('/history/stats'),
  cleanup: () =>
    api.post('/history/cleanup'),
};

// Error handling helper
export const handleAPIError = (error: unknown): string => {
  if (axios.isAxiosError(error)) {
    return error.response?.data?.message || error.message || 'An error occurred';
  }
  return 'An unexpected error occurred';
};

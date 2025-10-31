// API type definitions matching Go backend models

export interface Episode {
  id: string;
  title: string;
  description: string;
  duration: number;
  size: number;
  status: 'new' | 'queued' | 'downloading' | 'downloaded' | 'error' | 'cleaned' | 'blocked' | 'ignored';
  pub_date: string;
  file_url: string;
  thumbnail: string;
  feed_id: string;
  feed_title: string;
  video_url: string;
  error: string;
}

export interface EpisodeListResponse {
  episodes: Episode[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface Feed {
  id: string;
  url: string;
  title: string;
  description: string;
  episode_count: number;
  last_update: string;
  status: string;
  configuration: FeedConfig;
  author: string;
  cover_art: string;
  provider: string;
  format: string;
  quality: string;
}

export interface FeedConfig {
  update_period: string;
  cron_schedule: string;
  quality: string;
  format: string;
  page_size: number;
  max_height: number;
  cleanup_keep: number;
  playlist_sort: string;
  private_feed: boolean;
  opml: boolean;
  filters: Filters;
  custom: Custom;
}

export interface Filters {
  title?: string;
  not_title?: string;
  description?: string;
  not_description?: string;
  min_duration?: number;
  max_duration?: number;
  max_age?: number;
  min_age?: number;
}

export interface Custom {
  cover_art?: string;
  cover_art_quality?: string;
  category?: string;
  subcategories?: string[];
  explicit: boolean;
  lang?: string;
  author?: string;
  title?: string;
  description?: string;
  owner_name?: string;
  owner_email?: string;
  link?: string;
}

export interface AppConfig {
  server: ServerConfig;
  storage: StorageConfig;
  feeds: Record<string, FeedConfig>;
  database: DatabaseConfig;
  downloader: DownloaderConfig;
  tokens: TokensConfig;
  history: HistoryConfig;
}

export interface HistoryConfig {
  enabled: boolean;
  retention_days: number;
  max_entries: number;
}

export interface TokensConfig {
  youtube?: string[];
  vimeo?: string[];
  soundcloud?: string[];
  twitch?: string[];
}

export interface ServerConfig {
  hostname: string;
  port: number;
  frontend_port: number;
  bind_address: string;
  tls: boolean;
  certificate_path?: string;
  key_file_path?: string;
  path: string;
  web_ui: boolean;
  basic_auth?: BasicAuthConfig;
}

export interface BasicAuthConfig {
  enabled: boolean;
  username: string;
  password: string;
}

export interface StorageConfig {
  type: string;
  local?: LocalStorageConfig;
  s3?: S3StorageConfig;
}

export interface LocalStorageConfig {
  data_dir: string;
}

export interface S3StorageConfig {
  endpoint_url: string;
  region: string;
  bucket: string;
  prefix?: string;
  access_key?: string;
  secret_key?: string;
}

export interface DatabaseConfig {
  dir: string;
}

export interface DownloaderConfig {
  self_update: boolean;
  update_channel?: string;
  update_version?: string;
  timeout: string;
  ytdl_version?: string;
}

// History types
export type JobType = 'feed_update' | 'episode_retry' | 'episode_delete' | 'episode_block';
export type JobStatus = 'running' | 'success' | 'failed' | 'partial';
export type TriggerType = 'scheduled' | 'manual';

export interface EpisodeDetail {
  id: string;
  title: string;
  status: string;
  error?: string;
  size?: number;
  duration?: number;
}

export interface JobStatistics {
  episodes_queued: number;
  episodes_downloaded: number;
  episodes_failed: number;
  episodes_ignored: number;
  bytes_downloaded: number;
  episode_details?: EpisodeDetail[];
}

export interface HistoryEntry {
  id: string;
  job_type: JobType;
  feed_id: string;
  feed_title: string;
  episode_id: string;
  episode_title: string;
  start_time: string;
  end_time?: string;
  duration: number;
  status: JobStatus;
  trigger_type: TriggerType;
  statistics: JobStatistics;
  error: string;
}

export interface HistoryListResponse {
  entries: HistoryEntry[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface HistoryStatsResponse {
  count: number;
  oldest_entry?: HistoryEntry;
}

export interface HistoryFilters {
  feed_id?: string;
  job_type?: JobType;
  status?: JobStatus;
  search?: string;
  start_date?: string;
  end_date?: string;
}

import React, { useEffect, useState } from 'react';
import { useHistoryStore } from '../stores/useHistoryStore';
import { useFeedsStore } from '../stores/useFeedsStore';
import { Card } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Search, Trash2, Loader2, ChevronLeft, ChevronRight, RefreshCw, AlertCircle, Clock, CheckCircle, XCircle, Activity, ChevronDown, ChevronUp } from 'lucide-react';
import type { HistoryEntry, JobType, JobStatus, EpisodeDetail } from '../types/api';

export const History: React.FC = () => {
  const { historyData, loading, error, loadHistory, loadStats, deleteHistory, deleteAllHistory, cleanup } = useHistoryStore();
  const { feeds, loadFeeds } = useFeedsStore();
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [selectedFeed, setSelectedFeed] = useState<string>('');
  const [selectedJobType, setSelectedJobType] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadFeeds();
    loadStats();
  }, [loadFeeds, loadStats]);

  useEffect(() => {
    loadHistory({
      page: page + 1,
      page_size: pageSize,
      search,
      feed_id: selectedFeed || undefined,
      job_type: selectedJobType as JobType | undefined,
      status: selectedStatus as JobStatus | undefined,
    });
  }, [page, pageSize, search, selectedFeed, selectedJobType, selectedStatus, loadHistory]);

  const handleDelete = async (id: string) => {
    if (confirm('Are you sure you want to delete this history entry?')) {
      await deleteHistory(id);
      loadStats();
    }
  };

  const handleDeleteAll = async () => {
    if (confirm('Are you sure you want to delete ALL history? This action cannot be undone.')) {
      await deleteAllHistory();
      loadStats();
    }
  };

  const handleCleanup = async () => {
    if (confirm('Clean up old history entries based on retention policy?')) {
      await cleanup();
      loadStats();
    }
  };

  const toggleRow = (id: string) => {
    setExpandedRows(prev => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  const getStatusIcon = (status: JobStatus) => {
    switch (status) {
      case 'running':
        return <Activity className="h-4 w-4 text-blue-500 animate-pulse" />;
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'partial':
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      default:
        return null;
    }
  };

  const getStatusBadgeClass = (status: JobStatus) => {
    switch (status) {
      case 'running':
        return 'bg-blue-100 text-blue-800';
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'partial':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getJobTypeLabel = (jobType: JobType) => {
    switch (jobType) {
      case 'feed_update':
        return 'Feed Update';
      case 'episode_retry':
        return 'Episode Retry';
      case 'episode_delete':
        return 'Episode Delete';
      case 'episode_block':
        return 'Episode Block';
      default:
        return jobType;
    }
  };

  const formatDuration = (durationNs: number) => {
    const seconds = Math.floor(durationNs / 1000000000);
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">History</h1>
          <p className="text-gray-500">
            Track feed updates, episode operations, and job statistics
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleCleanup} disabled={loading} title="Remove history entries based on retention policy settings">
            <RefreshCw className="mr-2 h-4 w-4" />
            Cleanup Old Entries
          </Button>
          <Button variant="destructive" onClick={handleDeleteAll} disabled={loading} title="Permanently delete all history entries">
            <Trash2 className="mr-2 h-4 w-4" />
            Delete All
          </Button>
        </div>
      </div>

      <Card className="p-4">
        <div className="flex flex-col md:flex-row gap-4 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search history..."
              className="w-full pl-10 pr-4 py-2 border rounded-lg"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(0);
              }}
            />
          </div>

          <select
            className="px-4 py-2 border rounded-lg"
            value={selectedFeed}
            onChange={(e) => {
              setSelectedFeed(e.target.value);
              setPage(0);
            }}
          >
            <option value="">All Feeds</option>
            {feeds.map((feed) => (
              <option key={feed.id} value={feed.id}>
                {feed.title}
              </option>
            ))}
          </select>

          <select
            className="px-4 py-2 border rounded-lg"
            value={selectedJobType}
            onChange={(e) => {
              setSelectedJobType(e.target.value);
              setPage(0);
            }}
          >
            <option value="">All Job Types</option>
            <option value="feed_update">Feed Update</option>
            <option value="episode_retry">Episode Retry</option>
            <option value="episode_delete">Episode Delete</option>
            <option value="episode_block">Episode Block</option>
          </select>

          <select
            className="px-4 py-2 border rounded-lg"
            value={selectedStatus}
            onChange={(e) => {
              setSelectedStatus(e.target.value);
              setPage(0);
            }}
          >
            <option value="">All Statuses</option>
            <option value="running">Running</option>
            <option value="success">Success</option>
            <option value="failed">Failed</option>
            <option value="partial">Partial</option>
          </select>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
            <AlertCircle className="h-5 w-5" />
            <span>{error}</span>
          </div>
        )}

        {loading && !historyData ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 w-10"></th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Job Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Feed
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Episode
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Started
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Duration
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Statistics
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {!historyData || !historyData.entries || historyData.entries.length === 0 ? (
                  <tr>
                    <td colSpan={9} className="px-6 py-12 text-center text-gray-500">
                      <Clock className="h-12 w-12 mx-auto mb-4 opacity-50" />
                      <p>No history entries found</p>
                    </td>
                  </tr>
                ) : (
                  historyData.entries.map((entry: HistoryEntry) => {
                    const isExpanded = expandedRows.has(entry.id);
                    const hasEpisodes = entry.statistics?.episode_details && entry.statistics.episode_details.length > 0;

                    return (
                      <React.Fragment key={entry.id}>
                        <tr className="hover:bg-gray-50">
                          <td className="px-4 py-4">
                            <button
                              onClick={() => toggleRow(entry.id)}
                              className="p-1 hover:bg-gray-200 rounded transition-colors"
                              title={isExpanded ? "Collapse episodes" : "Expand episodes"}
                            >
                              {isExpanded ? (
                                <ChevronUp className="w-4 h-4 text-gray-600" />
                              ) : (
                                <ChevronDown className="w-4 h-4 text-gray-600" />
                              )}
                            </button>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex items-center gap-2">
                              {getStatusIcon(entry.status)}
                              <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(entry.status)}`}>
                                {entry.status}
                              </span>
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-gray-700">{getJobTypeLabel(entry.job_type)}</div>
                            <div className="text-xs text-gray-500">{entry.trigger_type}</div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm font-medium text-gray-900">{entry.feed_title}</div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-gray-700">
                              {entry.episode_title || '-'}
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-gray-700">{formatDate(entry.start_time)}</div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-gray-700">
                              {entry.end_time ? formatDuration(entry.duration) : '-'}
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            {entry.job_type === 'feed_update' && entry.statistics ? (
                              <div className="text-sm space-y-1">
                                <div className="flex items-center gap-3">
                                  <span className="text-blue-600" title="Queued: Episodes queued for download">
                                    Q: {entry.statistics.episodes_queued}
                                  </span>
                                  <span className="text-green-600" title="Downloaded: Episodes successfully downloaded">
                                    D: {entry.statistics.episodes_downloaded}
                                  </span>
                                  {entry.statistics.episodes_failed > 0 && (
                                    <span className="text-red-600" title="Failed: Episodes that failed to download">
                                      F: {entry.statistics.episodes_failed}
                                    </span>
                                  )}
                                  {entry.statistics.episodes_ignored > 0 && (
                                    <span className="text-gray-600" title="Ignored: Episodes skipped (e.g., too short)">
                                      I: {entry.statistics.episodes_ignored}
                                    </span>
                                  )}
                                </div>
                                {entry.statistics.bytes_downloaded > 0 && (
                                  <div className="text-xs text-gray-500">{formatBytes(entry.statistics.bytes_downloaded)}</div>
                                )}
                              </div>
                            ) : entry.error ? (
                              <div className="group relative inline-block">
                                <AlertCircle className="w-4 h-4 text-red-600 cursor-help" />
                                <div className="absolute left-0 bottom-full mb-2 hidden group-hover:block w-64 p-3 bg-gray-900 text-white text-xs rounded-lg shadow-lg z-10">
                                  <div className="font-semibold mb-1">Error Details:</div>
                                  <div className="break-words">{entry.error}</div>
                                  <div className="absolute left-4 top-full w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
                                </div>
                              </div>
                            ) : (
                              <span className="text-sm text-gray-400">-</span>
                            )}
                          </td>
                          <td className="px-6 py-4">
                            <button
                              onClick={() => handleDelete(entry.id)}
                              disabled={loading}
                              className="p-1.5 text-red-600 hover:bg-red-50 rounded-md transition-colors disabled:opacity-50"
                              title="Delete entry"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </td>
                        </tr>

                        {/* Expanded episode details row */}
                        {isExpanded && (
                          <tr key={`${entry.id}-details`} className="bg-gray-50">
                            <td colSpan={9} className="px-6 py-4">
                              <div className="space-y-4">
                                <h4 className="font-semibold text-sm text-gray-700">Episodes ({entry.statistics.episode_details?.length || 0})</h4>
                                {hasEpisodes ? (
                                  <div className="grid grid-cols-1 gap-3">
                                    {entry.statistics.episode_details?.map((episode: EpisodeDetail) => (
                                      <div
                                        key={episode.id}
                                        className="bg-white rounded-lg border border-gray-200 p-3 hover:shadow-sm transition-shadow"
                                      >
                                        <div className="flex items-start justify-between gap-3">
                                          <div className="flex-1 min-w-0">
                                            <div className="flex items-center gap-2 mb-1">
                                              <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                                                episode.status === 'downloaded' ? 'bg-green-100 text-green-800' :
                                                episode.status === 'error' ? 'bg-red-100 text-red-800' :
                                                episode.status === 'ignored' ? 'bg-gray-100 text-gray-800' :
                                                episode.status === 'queued' ? 'bg-blue-100 text-blue-800' :
                                                episode.status === 'new' ? 'bg-yellow-100 text-yellow-800' :
                                                'bg-gray-100 text-gray-800'
                                              }`}>
                                                {episode.status}
                                              </span>
                                              {episode.size && episode.size > 0 && (
                                                <span className="text-xs text-gray-500">{formatBytes(episode.size)}</span>
                                              )}
                                              {episode.duration && episode.duration > 0 && (
                                                <span className="text-xs text-gray-500">{formatDuration(episode.duration * 1000000000)}</span>
                                              )}
                                            </div>
                                            <p className="text-sm text-gray-900 truncate" title={episode.title}>
                                              {episode.title}
                                            </p>
                                            {episode.error && (
                                              <p className="text-xs text-red-600 mt-1">
                                                Error: {episode.error}
                                              </p>
                                            )}
                                          </div>
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                ) : (
                                  <div className="text-center py-8 text-gray-500">
                                    <p className="text-sm">No episode data available for this job</p>
                                    <p className="text-xs mt-1">No episodes were queued, downloaded, or ignored during this update</p>
                                  </div>
                                )}
                              </div>
                            </td>
                          </tr>
                        )}
                      </React.Fragment>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        )}

        <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="text-sm text-gray-700">
              Showing <span className="font-medium">{historyData && historyData.total > 0 ? page * pageSize + 1 : 0}</span> to{' '}
              <span className="font-medium">{Math.min((page + 1) * pageSize, historyData?.total || 0)}</span> of{' '}
              <span className="font-medium">{historyData?.total || 0}</span> results
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-700">Per page:</span>
              <select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setPage(0);
                }}
                className="px-3 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-sm"
              >
                <option value="10">10</option>
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
              </select>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(0, p - 1))}
              disabled={page === 0 || loading}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <span className="text-sm text-gray-700">
              Page {page + 1} of {historyData?.total_pages || 1}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => p + 1)}
              disabled={page >= (historyData?.total_pages || 1) - 1 || loading}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
};

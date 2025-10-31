import { useEffect, useState } from 'react';
import { useEpisodesStore } from '../stores/useEpisodesStore';
import { useFeedsStore } from '../stores/useFeedsStore';
import { Card } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Search, Play, Trash2, Loader2, X, ChevronLeft, ChevronRight, ExternalLink, AlertCircle, RotateCw, Ban } from 'lucide-react';
import { cn } from '../lib/utils';

export const Episodes: React.FC = () => {
  const { episodesData, loading, error, loadEpisodes, deleteEpisode, retryEpisode, blockEpisode } = useEpisodesStore();
  const { feeds, loadFeeds } = useFeedsStore();
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [currentAudio, setCurrentAudio] = useState<string | null>(null);
  const [selectedEpisodes, setSelectedEpisodes] = useState<Set<string>>(new Set());
  const [showIgnored, setShowIgnored] = useState(false);
  const [selectedFeed, setSelectedFeed] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [dateFilter, setDateFilter] = useState<string>('');
  const [customDateStart, setCustomDateStart] = useState<string>('');
  const [customDateEnd, setCustomDateEnd] = useState<string>('');

  useEffect(() => {
    loadFeeds();
  }, [loadFeeds]);

  useEffect(() => {
    loadEpisodes({
      page: page + 1,
      page_size: pageSize,
      search,
      show_ignored: showIgnored,
      feed_id: selectedFeed || undefined,
      status: selectedStatus || undefined,
      date_filter: dateFilter === 'custom' ? undefined : (dateFilter || undefined),
      date_start: dateFilter === 'custom' && customDateStart ? customDateStart : undefined,
      date_end: dateFilter === 'custom' && customDateEnd ? customDateEnd : undefined,
    });
  }, [page, pageSize, search, showIgnored, selectedFeed, selectedStatus, dateFilter, customDateStart, customDateEnd, loadEpisodes]);

  const reloadEpisodes = () => {
    loadEpisodes({
      page: page + 1,
      page_size: pageSize,
      search,
      show_ignored: showIgnored,
      feed_id: selectedFeed || undefined,
      status: selectedStatus || undefined,
      date_filter: dateFilter === 'custom' ? undefined : (dateFilter || undefined),
      date_start: dateFilter === 'custom' && customDateStart ? customDateStart : undefined,
      date_end: dateFilter === 'custom' && customDateEnd ? customDateEnd : undefined,
    });
  };

  const handleDelete = async (feedId: string, episodeId: string) => {
    if (confirm('Are you sure you want to delete this episode?')) {
      await deleteEpisode(feedId, episodeId);
      reloadEpisodes();
    }
  };

  const handleRetry = async (feedId: string, episodeId: string) => {
    try {
      await retryEpisode(feedId, episodeId);
      // Reload the episodes list to show updated status
      setTimeout(() => {
        reloadEpisodes();
      }, 1000);
    } catch (err) {
      console.error('Failed to retry episode:', err);
    }
  };

  const handleBlock = async (feedId: string, episodeId: string) => {
    if (confirm('Are you sure you want to block this episode? It will never be re-downloaded.')) {
      try {
        await blockEpisode(feedId, episodeId);
      } catch (err) {
        console.error('Failed to block episode:', err);
      }
    }
  };

  const handleBulkDelete = async () => {
    if (selectedEpisodes.size === 0) return;

    if (confirm(`Are you sure you want to delete ${selectedEpisodes.size} selected episode(s)?`)) {
      // Delete each selected episode
      const promises = Array.from(selectedEpisodes).map(async (episodeKey) => {
        const [feedId, episodeId] = episodeKey.split('|');
        await deleteEpisode(feedId, episodeId);
      });

      await Promise.all(promises);
      setSelectedEpisodes(new Set());
      reloadEpisodes();
    }
  };

  const toggleEpisodeSelection = (feedId: string, episodeId: string) => {
    const key = `${feedId}|${episodeId}`;
    const newSelection = new Set(selectedEpisodes);
    if (newSelection.has(key)) {
      newSelection.delete(key);
    } else {
      newSelection.add(key);
    }
    setSelectedEpisodes(newSelection);
  };

  const toggleSelectAll = () => {
    if (selectedEpisodes.size === (episodesData?.episodes || []).length) {
      setSelectedEpisodes(new Set());
    } else {
      const allKeys = (episodesData?.episodes || []).map(ep => `${ep.feed_id}|${ep.id}`);
      setSelectedEpisodes(new Set(allKeys));
    }
  };

  const handlePlay = (fileUrl: string) => {
    setCurrentAudio(fileUrl);
  };

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m`;
  };

  const formatSize = (bytes: number) => {
    const mb = (bytes / (1024 * 1024)).toFixed(1);
    return `${mb} MB`;
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - date.getTime());
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
    return date.toLocaleDateString();
  };

  const truncateDescription = (text: string, maxLength: number = 100) => {
    if (!text || text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      downloaded: 'bg-green-100 text-green-700',
      new: 'bg-yellow-100 text-yellow-700',
      queued: 'bg-yellow-100 text-yellow-700',
      downloading: 'bg-blue-100 text-blue-700',
      error: 'bg-red-100 text-red-700',
      cleaned: 'bg-gray-100 text-gray-700',
      blocked: 'bg-purple-100 text-purple-700',
      ignored: 'bg-orange-100 text-orange-700',
    };
    return colors[status] || 'bg-gray-100 text-gray-700';
  };

  const totalPages = Math.ceil((episodesData?.total || 0) / pageSize);

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Episodes</h1>
        <p className="text-gray-600 mt-1">Browse and manage all episodes</p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800 mb-4">
          {error}
        </div>
      )}

      {/* Search and Filters */}
      <div className="mb-6 space-y-4">
        <div className="flex gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search episodes..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <select
            value={selectedFeed}
            onChange={(e) => {
              setSelectedFeed(e.target.value);
              setPage(0);
            }}
            className="pl-4 pr-10 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
          >
            <option value="">All Feeds</option>
            {feeds.map((feed) => (
              <option key={feed.id} value={feed.id}>
                {feed.title}
              </option>
            ))}
          </select>
          <select
            value={selectedStatus}
            onChange={(e) => {
              setSelectedStatus(e.target.value);
              setPage(0);
            }}
            className="pl-4 pr-10 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
          >
            <option value="">All Statuses</option>
            <option value="downloaded">Downloaded</option>
            <option value="new">New</option>
            <option value="queued">Queued</option>
            <option value="downloading">Downloading</option>
            <option value="error">Error</option>
            <option value="cleaned">Cleaned</option>
            <option value="blocked">Blocked</option>
            <option value="ignored">Ignored</option>
          </select>
          <select
            value={dateFilter}
            onChange={(e) => {
              setDateFilter(e.target.value);
              if (e.target.value !== 'custom') {
                setCustomDateStart('');
                setCustomDateEnd('');
              }
              setPage(0);
            }}
            className="pl-4 pr-10 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
          >
            <option value="">All Time</option>
            <option value="today">Today</option>
            <option value="yesterday">Yesterday</option>
            <option value="week">Last 7 Days</option>
            <option value="month">Last Month</option>
            <option value="year">Last Year</option>
            <option value="custom">Custom Range</option>
          </select>
          {dateFilter === 'custom' && (
            <>
              <input
                type="date"
                value={customDateStart}
                onChange={(e) => {
                  setCustomDateStart(e.target.value);
                  setPage(0);
                }}
                className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
                placeholder="Start date"
              />
              <input
                type="date"
                value={customDateEnd}
                onChange={(e) => {
                  setCustomDateEnd(e.target.value);
                  setPage(0);
                }}
                className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
                placeholder="End date"
              />
            </>
          )}
          <label className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg cursor-pointer hover:bg-gray-50 whitespace-nowrap">
            <input
              type="checkbox"
              checked={showIgnored}
              onChange={(e) => setShowIgnored(e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
            />
            <span className="text-sm text-gray-700">Show ignored</span>
          </label>
          {selectedEpisodes.size > 0 && (
            <Button
              variant="outline"
              onClick={handleBulkDelete}
              className="flex items-center gap-2"
              title="Permanently delete all selected episodes and their files"
            >
              <Trash2 className="w-4 h-4" />
              Delete Selected ({selectedEpisodes.size})
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      {loading && !episodesData ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
        </div>
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={selectedEpisodes.size > 0 && selectedEpisodes.size === (episodesData?.episodes || []).length}
                      onChange={toggleSelectAll}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Title
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Feed
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Published
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Duration
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Size
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {(episodesData?.episodes || []).map((episode) => {
                  const episodeKey = `${episode.feed_id}|${episode.id}`;
                  const isSelected = selectedEpisodes.has(episodeKey);

                  return (
                    <tr key={episode.id} className={cn("hover:bg-gray-50", isSelected && "bg-blue-50")}>
                      <td className="px-4 py-4">
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleEpisodeSelection(episode.feed_id, episode.id)}
                          className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                        />
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm font-medium text-gray-900">{episode.title}</div>
                      </td>
                    <td className="px-6 py-4 max-w-xs">
                      <div className="text-sm text-gray-600" title={episode.description}>
                        {truncateDescription(episode.description)}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-700">{episode.feed_title}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-700" title={episode.pub_date ? new Date(episode.pub_date).toLocaleString() : ''}>
                        {formatDate(episode.pub_date)}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <span className={cn(
                          "px-2.5 py-1 rounded-full text-xs font-medium flex items-center gap-1",
                          getStatusColor(episode.status)
                        )}>
                          {episode.status === 'downloading' && (
                            <Loader2 className="w-3 h-3 animate-spin" />
                          )}
                          {episode.status}
                        </span>
                        {episode.status === 'error' && episode.error && (
                          <div className="group relative">
                            <AlertCircle className="w-4 h-4 text-red-600 cursor-help" />
                            <div className="absolute left-0 bottom-full mb-2 hidden group-hover:block w-64 p-3 bg-gray-900 text-white text-xs rounded-lg shadow-lg z-10">
                              <div className="font-semibold mb-1">Error Details:</div>
                              <div className="break-words">{episode.error}</div>
                              <div className="absolute left-4 top-full w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
                            </div>
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-700">{formatDuration(episode.duration)}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-700">{formatSize(episode.size)}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        {episode.video_url && (
                          <a
                            href={episode.video_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="p-1.5 text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
                            title="View original video"
                          >
                            <ExternalLink className="w-4 h-4" />
                          </a>
                        )}
                        {episode.file_url && (
                          <button
                            onClick={() => handlePlay(episode.file_url)}
                            className="p-1.5 text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
                            title="Play"
                          >
                            <Play className="w-4 h-4" />
                          </button>
                        )}
                        {(episode.status === 'error' || episode.status === 'new') && (
                          <button
                            onClick={() => handleRetry(episode.feed_id, episode.id)}
                            className="p-1.5 text-orange-600 hover:bg-orange-50 rounded-md transition-colors"
                            title="Retry download"
                          >
                            <RotateCw className="w-4 h-4" />
                          </button>
                        )}
                        {episode.status !== 'blocked' && (
                          <button
                            onClick={() => handleBlock(episode.feed_id, episode.id)}
                            className="p-1.5 text-purple-600 hover:bg-purple-50 rounded-md transition-colors"
                            title="Block episode from being re-downloaded"
                          >
                            <Ban className="w-4 h-4" />
                          </button>
                        )}
                        <button
                          onClick={() => handleDelete(episode.feed_id, episode.id)}
                          className="p-1.5 text-red-600 hover:bg-red-50 rounded-md transition-colors"
                          title="Delete"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="text-sm text-gray-700">
                Showing <span className="font-medium">{page * pageSize + 1}</span> to{' '}
                <span className="font-medium">{Math.min((page + 1) * pageSize, episodesData?.total || 0)}</span> of{' '}
                <span className="font-medium">{episodesData?.total || 0}</span> results
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
                disabled={page === 0}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <span className="text-sm text-gray-700">
                Page {page + 1} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => p + 1)}
                disabled={page >= totalPages - 1}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Audio Player */}
      {currentAudio && (
        <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg p-4 z-50">
          <div className="max-w-7xl mx-auto flex items-center gap-4">
            <audio
              controls
              autoPlay
              src={currentAudio}
              className="flex-1"
              onEnded={() => setCurrentAudio(null)}
            />
            <button
              onClick={() => setCurrentAudio(null)}
              className="p-2 text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

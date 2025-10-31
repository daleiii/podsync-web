import { useEffect, useState } from 'react';
import { Card } from './ui/card';
import { Loader2, Download } from 'lucide-react';

interface EpisodeProgress {
  feed_id: string;
  episode_id: string;
  episode_title: string;
  stage: string;
  percent: number;
  downloaded: number;
  total: number;
  speed: string;
  start_time: string;
  last_update: string;
}

interface FeedProgress {
  feed_id: string;
  total_episodes: number;
  completed_count: number;
  downloading_count: number;
  queued_count: number;
  overall_percent: number;
  start_time: string;
}

interface ProgressData {
  feeds: Record<string, FeedProgress>;
  episodes: EpisodeProgress[];
}

export const DownloadProgress: React.FC = () => {
  const [progressData, setProgressData] = useState<ProgressData | null>(null);

  useEffect(() => {
    // Connect to SSE stream
    const es = new EventSource('/api/v1/progress/stream');

    es.onmessage = (event) => {
      try {
        const data: ProgressData = JSON.parse(event.data);
        setProgressData(data);
      } catch (err) {
        console.error('Failed to parse progress data:', err);
      }
    };

    es.onerror = (err) => {
      console.error('SSE error:', err);
      // Connection will auto-reconnect, just log the error
    };

    return () => {
      es.close();
    };
  }, []); // Empty dependency array - only run once on mount

  // Don't render if no active downloads
  if (!progressData || Object.keys(progressData.feeds).length === 0) {
    return null;
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatSpeed = (speed: string) => {
    if (!speed) return '';
    return speed;
  };

  return (
    <Card className="mb-6 p-6">
      <div className="flex items-center gap-2 mb-4">
        <Download className="w-5 h-5 text-blue-600" />
        <h3 className="text-lg font-semibold text-gray-900">Active Downloads</h3>
      </div>

      {Object.entries(progressData.feeds).map(([feedId, feedProgress]) => (
        <div key={feedId} className="mb-6 last:mb-0">
          {/* Feed-level progress */}
          <div className="mb-3">
            <div className="flex items-center justify-between mb-2">
              <div className="text-sm font-medium text-gray-700">
                Feed: {feedId}
              </div>
              <div className="text-sm text-gray-600">
                {feedProgress.completed_count} of {feedProgress.total_episodes} episodes
                {feedProgress.downloading_count > 0 && (
                  <span className="ml-2 text-blue-600">
                    ({feedProgress.downloading_count} downloading)
                  </span>
                )}
              </div>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${feedProgress.overall_percent}%` }}
              />
            </div>
            <div className="text-xs text-gray-500 mt-1">
              {feedProgress.overall_percent.toFixed(1)}% complete
            </div>
          </div>

          {/* Episode-level progress */}
          {progressData.episodes
            .filter((ep) => ep.feed_id === feedId)
            .map((episode) => (
              <div
                key={episode.episode_id}
                className="ml-4 p-3 bg-gray-50 rounded-lg mb-2"
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium text-gray-900 truncate">
                      {episode.episode_title}
                    </div>
                    <div className="text-xs text-gray-500">
                      {episode.stage === 'downloading' && 'Downloading'}
                      {episode.stage === 'encoding' && 'Encoding'}
                      {episode.stage === 'saving' && 'Saving'}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    {episode.speed && (
                      <span className="text-xs text-gray-600">
                        {formatSpeed(episode.speed)}
                      </span>
                    )}
                    <Loader2 className="w-4 h-4 animate-spin text-blue-600" />
                  </div>
                </div>

                {episode.stage === 'downloading' && episode.total > 0 && (
                  <>
                    <div className="w-full bg-gray-200 rounded-full h-1.5 mb-1">
                      <div
                        className="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
                        style={{ width: `${episode.percent}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-xs text-gray-500">
                      <span>
                        {formatBytes(episode.downloaded)} / {formatBytes(episode.total)}
                      </span>
                      <span>{episode.percent.toFixed(1)}%</span>
                    </div>
                  </>
                )}

                {episode.stage === 'encoding' && (
                  <div className="text-xs text-gray-500">
                    Post-processing video/audio...
                  </div>
                )}
              </div>
            ))}
        </div>
      ))}
    </Card>
  );
};

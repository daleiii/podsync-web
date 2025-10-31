import { useEffect } from 'react';
import { useFeedsStore } from '../stores/useFeedsStore';
import { useConfigStore } from '../stores/useConfigStore';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Rss, Radio, TrendingUp, Loader2, FileText } from 'lucide-react';
import { DownloadProgress } from '../components/DownloadProgress';

export const Dashboard: React.FC = () => {
  const { feeds, loading, error, loadFeeds } = useFeedsStore();
  const { getBackendURL } = useConfigStore();

  useEffect(() => {
    loadFeeds();
  }, [loadFeeds]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
        {error}
      </div>
    );
  }

  const totalEpisodes = feeds.reduce((sum, feed) => sum + feed.episode_count, 0);
  const activeFeeds = feeds.filter((f) => f.status === 'active').length;
  const hasOpmlFeeds = feeds.some((f) => f.configuration?.opml);

  return (
    <div>
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
            <p className="text-gray-600 mt-1">Overview of your podcast feeds</p>
          </div>
          {hasOpmlFeeds && (
            <a
              href={`${getBackendURL()}/podsync.opml`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              title="Export all feeds as OPML file for importing into podcast clients"
            >
              <FileText className="w-4 h-4" />
              Download OPML
            </a>
          )}
        </div>
      </div>

      {/* Download Progress */}
      <DownloadProgress />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Card className="border-l-4 border-l-blue-500">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Feeds</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{feeds.length}</p>
              </div>
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                <Rss className="w-6 h-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-l-4 border-l-green-500">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Feeds</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{activeFeeds}</p>
              </div>
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                <TrendingUp className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-l-4 border-l-purple-500">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Episodes</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{totalEpisodes}</p>
              </div>
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
                <Radio className="w-6 h-6 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Feeds List */}
      <div>
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Your Feeds</h2>

        {feeds.length === 0 ? (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 text-center">
            <Rss className="w-12 h-12 text-blue-600 mx-auto mb-3" />
            <p className="text-blue-900 font-medium">No feeds configured yet</p>
            <p className="text-blue-700 text-sm mt-1">Add your first feed in the Feeds page</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {feeds.map((feed) => (
              <Card key={feed.id} className="hover:shadow-md transition-shadow">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="text-lg">{feed.title || feed.id}</CardTitle>
                      <p className="text-sm text-gray-600 mt-1">{feed.description}</p>
                    </div>
                    <span className="px-2.5 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded-full">
                      {feed.provider}
                    </span>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full">
                      {feed.episode_count} episodes
                    </span>
                    <span className="px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full">
                      {feed.format}
                    </span>
                    <span className="px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full">
                      {feed.quality}
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

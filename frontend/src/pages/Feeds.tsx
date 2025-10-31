import { useEffect, useState } from 'react';
import { useFeedsStore } from '../stores/useFeedsStore';
import { useConfigStore } from '../stores/useConfigStore';
import { Card, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Select } from '../components/ui/select';
import { Label } from '../components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from '../components/ui/dialog';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/tabs';
import { Plus, Pencil, Trash2, Loader2, Rss, RefreshCw, Folder } from 'lucide-react';
import type { Feed } from '../types/api';

interface FeedFormData {
  id: string;
  url: string;
  format: string;
  quality: string;
  update_period: string;
  cron_schedule: string;
  schedule_mode: 'simple' | 'advanced'; // Toggle between simple interval and advanced cron
  max_height: number;
  page_size: number;
  playlist_sort: string;
  opml: boolean;
  private_feed: boolean;
  cleanup_keep: number;
  // Custom format
  custom_format_youtube_dl: string;
  custom_format_extension: string;
  // Filters
  filter_title: string;
  filter_not_title: string;
  filter_description: string;
  filter_not_description: string;
  filter_min_duration: number;
  filter_max_duration: number;
  filter_min_age: number;
  filter_max_age: number;
  // Custom metadata
  custom_cover_art: string;
  custom_cover_art_quality: string;
  custom_category: string;
  custom_subcategories: string;
  custom_explicit: boolean;
  custom_lang: string;
  custom_author: string;
  custom_title: string;
  custom_description: string;
  custom_owner_name: string;
  custom_owner_email: string;
  custom_link: string;
  // Advanced settings
  youtube_dl_args: string;
  post_download_command: string;
  post_download_timeout: number;
  // Notification settings
  webhook_enabled: boolean;
  webhook_url: string;
  webhook_method: string;
  webhook_message: string;
  webhook_timeout: number;
}

export const Feeds: React.FC = () => {
  const { feeds, loading, error, loadFeeds, deleteFeed, refreshFeed } = useFeedsStore();
  const { getBackendURL } = useConfigStore();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingFeed, setEditingFeed] = useState<Feed | null>(null);
  const [formData, setFormData] = useState<FeedFormData>({
    id: '',
    url: '',
    format: 'video',
    quality: 'high',
    update_period: '12h',
    cron_schedule: '',
    schedule_mode: 'simple',
    max_height: 720,
    page_size: 50,
    playlist_sort: 'asc',
    opml: true,
    private_feed: false,
    cleanup_keep: 0,
    custom_format_youtube_dl: '',
    custom_format_extension: '',
    filter_title: '',
    filter_not_title: '',
    filter_description: '',
    filter_not_description: '',
    filter_min_duration: 0,
    filter_max_duration: 0,
    filter_min_age: 0,
    filter_max_age: 0,
    custom_cover_art: '',
    custom_cover_art_quality: 'high',
    custom_category: '',
    custom_subcategories: '',
    custom_explicit: false,
    custom_lang: 'en',
    custom_author: '',
    custom_title: '',
    custom_description: '',
    custom_owner_name: '',
    custom_owner_email: '',
    custom_link: '',
    youtube_dl_args: '',
    post_download_command: '',
    post_download_timeout: 120,
    webhook_enabled: false,
    webhook_url: '',
    webhook_method: 'POST',
    webhook_message: 'New episode: $EPISODE_TITLE',
    webhook_timeout: 30,
  });

  useEffect(() => {
    loadFeeds();
  }, [loadFeeds]);

  const handleAdd = () => {
    setEditingFeed(null);
    setFormData({
      id: '',
      url: '',
      format: 'video',
      quality: 'high',
      update_period: '12h',
      cron_schedule: '',
      schedule_mode: 'simple',
      max_height: 720,
      page_size: 50,
      playlist_sort: 'asc',
      opml: true,
      private_feed: false,
      cleanup_keep: 0,
      custom_format_youtube_dl: '',
      custom_format_extension: '',
      filter_title: '',
      filter_not_title: '',
      filter_description: '',
      filter_not_description: '',
      filter_min_duration: 0,
      filter_max_duration: 0,
      filter_min_age: 0,
      filter_max_age: 0,
      custom_cover_art: '',
      custom_cover_art_quality: 'high',
      custom_category: '',
      custom_subcategories: '',
      custom_explicit: false,
      custom_lang: 'en',
      custom_author: '',
      custom_title: '',
      custom_description: '',
      custom_owner_name: '',
      custom_owner_email: '',
      custom_link: '',
      youtube_dl_args: '',
      post_download_command: '',
      post_download_timeout: 120,
      webhook_enabled: false,
      webhook_url: '',
      webhook_method: 'POST',
      webhook_message: 'New episode: $EPISODE_TITLE',
      webhook_timeout: 30,
    });
    setDialogOpen(true);
  };

  const handleEdit = (feed: Feed) => {
    setEditingFeed(feed);
    const config = feed.configuration;
    // Determine schedule mode based on which field is populated
    const scheduleMode = config?.cron_schedule ? 'advanced' : 'simple';
    setFormData({
      id: feed.id,
      url: feed.url,
      format: feed.format || 'video',
      quality: feed.quality || 'high',
      update_period: config?.update_period || '12h',
      cron_schedule: config?.cron_schedule || '',
      schedule_mode: scheduleMode,
      max_height: config?.max_height || 720,
      page_size: config?.page_size || 50,
      playlist_sort: config?.playlist_sort || 'asc',
      opml: config?.opml ?? true,
      private_feed: config?.private_feed ?? false,
      cleanup_keep: config?.cleanup_keep || 0,
      custom_format_youtube_dl: (config as any)?.custom_format?.youtube_dl_format || '',
      custom_format_extension: (config as any)?.custom_format?.extension || '',
      filter_title: config?.filters?.title || '',
      filter_not_title: config?.filters?.not_title || '',
      filter_description: config?.filters?.description || '',
      filter_not_description: config?.filters?.not_description || '',
      filter_min_duration: config?.filters?.min_duration || 0,
      filter_max_duration: config?.filters?.max_duration || 0,
      filter_min_age: config?.filters?.min_age || 0,
      filter_max_age: config?.filters?.max_age || 0,
      custom_cover_art: config?.custom?.cover_art || '',
      custom_cover_art_quality: config?.custom?.cover_art_quality || 'high',
      custom_category: config?.custom?.category || '',
      custom_subcategories: config?.custom?.subcategories?.join(', ') || '',
      custom_explicit: config?.custom?.explicit ?? false,
      custom_lang: config?.custom?.lang || 'en',
      custom_author: config?.custom?.author || '',
      custom_title: config?.custom?.title || '',
      custom_description: config?.custom?.description || '',
      custom_owner_name: config?.custom?.owner_name || '',
      custom_owner_email: config?.custom?.owner_email || '',
      custom_link: config?.custom?.link || '',
      youtube_dl_args: (config as any)?.youtube_dl_args?.join(', ') || '',
      post_download_command: (config as any)?.post_episode_download?.[0]?.command?.join(' ') || '',
      post_download_timeout: (config as any)?.post_episode_download?.[0]?.timeout || 120,
      // Check if webhook is configured (looking for curl command pattern)
      webhook_enabled: (() => {
        const webhookHook = (config as any)?.post_episode_download?.find((hook: any) =>
          hook.command?.[0] === 'curl' && hook.command?.includes('-X')
        );
        return !!webhookHook;
      })(),
      webhook_url: (() => {
        const webhookHook = (config as any)?.post_episode_download?.find((hook: any) =>
          hook.command?.[0] === 'curl' && hook.command?.includes('-X')
        );
        return webhookHook?.command?.[webhookHook.command.length - 1] || '';
      })(),
      webhook_method: (() => {
        const webhookHook = (config as any)?.post_episode_download?.find((hook: any) =>
          hook.command?.[0] === 'curl' && hook.command?.includes('-X')
        );
        const methodIdx = webhookHook?.command?.indexOf('-X');
        return methodIdx !== undefined && methodIdx >= 0 ? webhookHook.command[methodIdx + 1] : 'POST';
      })(),
      webhook_message: (() => {
        const webhookHook = (config as any)?.post_episode_download?.find((hook: any) =>
          hook.command?.[0] === 'curl' && hook.command?.includes('-X')
        );
        const dataIdx = webhookHook?.command?.indexOf('-d');
        return dataIdx !== undefined && dataIdx >= 0 ? webhookHook.command[dataIdx + 1] : 'New episode: $EPISODE_TITLE';
      })(),
      webhook_timeout: (() => {
        const webhookHook = (config as any)?.post_episode_download?.find((hook: any) =>
          hook.command?.[0] === 'curl' && hook.command?.includes('-X')
        );
        return webhookHook?.timeout || 30;
      })(),
    });
    setDialogOpen(true);
  };

  const handleDelete = async (feedId: string) => {
    if (confirm(`Are you sure you want to delete feed "${feedId}"?\n\nThis will also delete all associated episodes and media files.`)) {
      try {
        await deleteFeed(feedId);
        loadFeeds(); // Reload the feeds list
      } catch (err) {
        console.error('Failed to delete feed:', err);
      }
    }
  };

  const handleRefresh = async (feedId: string) => {
    try {
      await refreshFeed(feedId);
      alert(`Feed "${feedId}" refresh triggered successfully! The feed will update in the background.`);
    } catch (err) {
      console.error('Failed to refresh feed:', err);
      alert(`Failed to refresh feed: ${err}`);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const payload = {
        id: formData.id,
        url: formData.url,
        config: {
          format: formData.format,
          quality: formData.quality,
          // Only send the relevant field based on schedule mode
          update_period: formData.schedule_mode === 'simple' ? formData.update_period : '',
          cron_schedule: formData.schedule_mode === 'advanced' ? formData.cron_schedule : '',
          max_height: formData.max_height,
          page_size: formData.page_size,
          playlist_sort: formData.playlist_sort,
          opml: formData.opml,
          private_feed: formData.private_feed,
          cleanup_keep: formData.cleanup_keep,
          custom_format: formData.format === 'custom' ? {
            youtube_dl_format: formData.custom_format_youtube_dl,
            extension: formData.custom_format_extension,
          } : undefined,
          filters: {
            title: formData.filter_title,
            not_title: formData.filter_not_title,
            description: formData.filter_description,
            not_description: formData.filter_not_description,
            min_duration: formData.filter_min_duration,
            max_duration: formData.filter_max_duration,
            min_age: formData.filter_min_age,
            max_age: formData.filter_max_age,
          },
          custom: {
            cover_art: formData.custom_cover_art,
            cover_art_quality: formData.custom_cover_art_quality,
            category: formData.custom_category,
            subcategories: formData.custom_subcategories.split(',').map(s => s.trim()).filter(s => s),
            explicit: formData.custom_explicit,
            lang: formData.custom_lang,
            author: formData.custom_author,
            title: formData.custom_title,
            description: formData.custom_description,
            owner_name: formData.custom_owner_name,
            owner_email: formData.custom_owner_email,
            link: formData.custom_link,
          },
          // Advanced settings
          youtube_dl_args: formData.youtube_dl_args ? formData.youtube_dl_args.split(',').map(s => s.trim()).filter(s => s) : undefined,
          post_episode_download: (() => {
            const hooks: any[] = [];

            // Add custom script hook if configured
            if (formData.post_download_command) {
              hooks.push({
                command: formData.post_download_command.split(' ').filter(s => s),
                timeout: formData.post_download_timeout,
              });
            }

            // Add webhook hook if enabled
            if (formData.webhook_enabled && formData.webhook_url) {
              hooks.push({
                command: [
                  'curl',
                  '-X',
                  formData.webhook_method,
                  '-d',
                  formData.webhook_message,
                  formData.webhook_url
                ],
                timeout: formData.webhook_timeout,
              });
            }

            return hooks.length > 0 ? hooks : undefined;
          })(),
        },
      };

      const url = editingFeed
        ? `/api/v1/feeds/${formData.id}`
        : '/api/v1/feeds';

      const method = editingFeed ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`Failed to ${editingFeed ? 'update' : 'create'} feed`);
      }

      const data = await response.json();
      alert(data.message || `Feed ${editingFeed ? 'updated' : 'created'} successfully!`);
      setDialogOpen(false);
      loadFeeds();
    } catch (error) {
      alert(`Error: ${(error as Error).message}`);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Feeds Management</h1>
          <p className="text-gray-600 mt-1">Manage your podcast feeds</p>
        </div>
        <Button onClick={handleAdd} title="Create a new podcast feed">
          <Plus className="w-4 h-4 mr-2" />
          Add Feed
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800 mb-6">
          {error}
        </div>
      )}

      {feeds.length === 0 ? (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-12 text-center">
          <Rss className="w-16 h-16 text-blue-600 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-blue-900 mb-2">No feeds yet</h3>
          <p className="text-blue-700 mb-4">Get started by adding your first podcast feed</p>
          <Button onClick={handleAdd} title="Create a new podcast feed">
            <Plus className="w-4 h-4 mr-2" />
            Add Your First Feed
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {feeds.map((feed) => (
            <Card key={feed.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <CardTitle className="text-xl mb-2">{feed.title || feed.id}</CardTitle>
                    <div className="flex flex-wrap gap-2 mb-3">
                      <span className="px-2.5 py-1 bg-blue-100 text-blue-700 text-xs font-medium rounded-full">
                        {feed.provider}
                      </span>
                      <span className="px-2.5 py-1 bg-gray-100 text-gray-700 text-xs rounded-full">
                        {feed.format}
                      </span>
                      <span className="px-2.5 py-1 bg-gray-100 text-gray-700 text-xs rounded-full">
                        {feed.quality}
                      </span>
                      <span className="px-2.5 py-1 bg-gray-100 text-gray-700 text-xs rounded-full">
                        {feed.episode_count} episodes
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{feed.description}</p>
                    <p className="text-xs text-gray-500 font-mono truncate mb-2">{feed.url}</p>
                    <div className="flex gap-3 text-xs">
                      <a
                        href={`${getBackendURL()}/${feed.id}.xml`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-blue-600 hover:text-blue-800 hover:underline"
                      >
                        <Rss className="w-3 h-3" />
                        RSS Feed
                      </a>
                      <a
                        href={`${getBackendURL()}/files/${feed.id}/`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-blue-600 hover:text-blue-800 hover:underline"
                      >
                        <Folder className="w-3 h-3" />
                        Episodes Folder
                      </a>
                    </div>
                  </div>
                  <div className="flex gap-2 ml-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRefresh(feed.id)}
                      title="Refresh feed - download new episodes"
                    >
                      <RefreshCw className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleEdit(feed)}
                      title="Edit feed settings and configuration"
                    >
                      <Pencil className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDelete(feed.id)}
                      title="Delete feed and all associated episodes"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}

      {/* Add/Edit Feed Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingFeed ? 'Edit Feed' : 'Add New Feed'}
            </DialogTitle>
            <DialogDescription>
              {editingFeed
                ? 'Update the feed configuration below'
                : 'Enter the details for your new podcast feed'}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit}>
            <Tabs defaultValue="basic">
              <TabsList>
                <TabsTrigger value="basic">Basic</TabsTrigger>
                <TabsTrigger value="schedule">Schedule & Cleanup</TabsTrigger>
                <TabsTrigger value="filters">Filters</TabsTrigger>
                <TabsTrigger value="metadata">Metadata</TabsTrigger>
                <TabsTrigger value="notifications">Notifications</TabsTrigger>
                <TabsTrigger value="advanced">Advanced</TabsTrigger>
              </TabsList>

              {/* Basic Tab */}
              <TabsContent value="basic">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="id">Feed ID</Label>
                    <Input
                      id="id"
                      value={formData.id}
                      onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                      placeholder="my-podcast"
                      required
                      disabled={!!editingFeed}
                    />
                    <p className="text-xs text-gray-500 mt-1">Unique identifier for this feed</p>
                  </div>

                  <div>
                    <Label htmlFor="url">Feed URL</Label>
                    <Input
                      id="url"
                      type="url"
                      value={formData.url}
                      onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                      placeholder="https://youtube.com/@channel"
                      required
                    />
                    <p className="text-xs text-gray-500 mt-1">YouTube, Vimeo, or SoundCloud URL</p>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="format">Format</Label>
                      <Select
                        id="format"
                        value={formData.format}
                        onChange={(e) => setFormData({ ...formData, format: e.target.value })}
                      >
                        <option value="video">Video</option>
                        <option value="audio">Audio</option>
                        <option value="custom">Custom</option>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="quality">Quality</Label>
                      <Select
                        id="quality"
                        value={formData.quality}
                        onChange={(e) => setFormData({ ...formData, quality: e.target.value })}
                      >
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                      </Select>
                    </div>
                  </div>

                  {formData.format === 'video' && (
                    <div>
                      <Label htmlFor="max_height">Max Height (pixels)</Label>
                      <Input
                        id="max_height"
                        type="number"
                        value={formData.max_height}
                        onChange={(e) => setFormData({ ...formData, max_height: parseInt(e.target.value) || 0 })}
                        placeholder="720"
                      />
                      <p className="text-xs text-gray-500 mt-1">Maximum video resolution (e.g., 720, 1080)</p>
                    </div>
                  )}

                  {formData.format === 'custom' && (
                    <>
                      <div>
                        <Label htmlFor="custom_format_youtube_dl">YouTube-DL Format String</Label>
                        <Input
                          id="custom_format_youtube_dl"
                          value={formData.custom_format_youtube_dl}
                          onChange={(e) => setFormData({ ...formData, custom_format_youtube_dl: e.target.value })}
                          placeholder="bestaudio[ext=m4a]"
                        />
                        <p className="text-xs text-gray-500 mt-1">Format string to pass to youtube-dl (e.g., &quot;bestaudio[ext=m4a]&quot;, &quot;best[height&lt;=720]&quot;)</p>
                      </div>

                      <div>
                        <Label htmlFor="custom_format_extension">File Extension</Label>
                        <Input
                          id="custom_format_extension"
                          value={formData.custom_format_extension}
                          onChange={(e) => setFormData({ ...formData, custom_format_extension: e.target.value })}
                          placeholder="m4a"
                        />
                        <p className="text-xs text-gray-500 mt-1">File extension for downloaded episodes (e.g., "m4a", "mp4", "mkv")</p>
                      </div>
                    </>
                  )}

                  <div>
                    <Label htmlFor="page_size">Episodes per Update</Label>
                    <Input
                      id="page_size"
                      type="number"
                      value={formData.page_size}
                      onChange={(e) => setFormData({ ...formData, page_size: parseInt(e.target.value) || 0 })}
                      placeholder="50"
                    />
                    <p className="text-xs text-gray-500 mt-1">Number of episodes to fetch per update</p>
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="opml"
                      checked={formData.opml}
                      onChange={(e) => setFormData({ ...formData, opml: e.target.checked })}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <Label htmlFor="opml" className="cursor-pointer">Include in OPML export</Label>
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="private_feed"
                      checked={formData.private_feed}
                      onChange={(e) => setFormData({ ...formData, private_feed: e.target.checked })}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <Label htmlFor="private_feed" className="cursor-pointer">Private feed (hide from indexers)</Label>
                  </div>
                </div>
              </TabsContent>

              {/* Schedule & Cleanup Tab */}
              <TabsContent value="schedule">
                <div className="space-y-4">
                  {/* Schedule Mode Toggle */}
                  <div className="space-y-3">
                    <Label>Schedule Mode</Label>
                    <div className="flex gap-3">
                      <button
                        type="button"
                        onClick={() => setFormData({ ...formData, schedule_mode: 'simple' })}
                        className={`flex-1 px-4 py-3 rounded-lg border-2 text-left transition-all ${
                          formData.schedule_mode === 'simple'
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <div className="font-semibold text-sm">Simple Interval</div>
                        <div className="text-xs text-gray-600 mt-1">
                          Updates at regular intervals. Runs immediately on startup.
                        </div>
                      </button>
                      <button
                        type="button"
                        onClick={() => setFormData({ ...formData, schedule_mode: 'advanced' })}
                        className={`flex-1 px-4 py-3 rounded-lg border-2 text-left transition-all ${
                          formData.schedule_mode === 'advanced'
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <div className="font-semibold text-sm">Advanced Cron</div>
                        <div className="text-xs text-gray-600 mt-1">
                          Precise scheduling with cron syntax. Waits for next scheduled time.
                        </div>
                      </button>
                    </div>
                  </div>

                  {/* Conditional Fields Based on Schedule Mode */}
                  {formData.schedule_mode === 'simple' ? (
                    <div>
                      <Label htmlFor="update_period">Update Period</Label>
                      <Input
                        id="update_period"
                        value={formData.update_period}
                        onChange={(e) => setFormData({ ...formData, update_period: e.target.value })}
                        placeholder="12h"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Examples: 12h, 6h30m, 24h, 30m
                      </p>
                      <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded-lg text-xs text-blue-900">
                        <strong>Note:</strong> Feed will update immediately on startup, then repeat every interval.
                      </div>
                    </div>
                  ) : (
                    <div>
                      <Label htmlFor="cron_schedule">Cron Schedule</Label>
                      <Input
                        id="cron_schedule"
                        value={formData.cron_schedule}
                        onChange={(e) => setFormData({ ...formData, cron_schedule: e.target.value })}
                        placeholder="@every 12h"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Examples: @every 12h, @daily, @hourly, or standard cron: 0 */6 * * *
                      </p>
                      <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded-lg text-xs text-blue-900">
                        <strong>Note:</strong> Feed waits for the next scheduled time (no immediate startup update).
                      </div>
                    </div>
                  )}

                  <div>
                    <Label htmlFor="playlist_sort">Playlist Sort Order</Label>
                    <Select
                      id="playlist_sort"
                      value={formData.playlist_sort}
                      onChange={(e) => setFormData({ ...formData, playlist_sort: e.target.value })}
                    >
                      <option value="asc">Ascending (oldest first)</option>
                      <option value="desc">Descending (newest first)</option>
                    </Select>
                  </div>

                  <div>
                    <Label htmlFor="cleanup_keep">Keep Last N Episodes</Label>
                    <Input
                      id="cleanup_keep"
                      type="number"
                      value={formData.cleanup_keep}
                      onChange={(e) => setFormData({ ...formData, cleanup_keep: parseInt(e.target.value) || 0 })}
                      placeholder="0"
                    />
                    <p className="text-xs text-gray-500 mt-1">0 = keep all episodes, N = keep last N episodes</p>
                  </div>
                </div>
              </TabsContent>

              {/* Filters Tab */}
              <TabsContent value="filters">
                <div className="space-y-4">
                  <p className="text-sm text-gray-600 mb-4">
                    Use regex patterns to filter episodes. Leave empty to include all.
                  </p>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="filter_title">Title Contains</Label>
                      <Input
                        id="filter_title"
                        value={formData.filter_title}
                        onChange={(e) => setFormData({ ...formData, filter_title: e.target.value })}
                        placeholder="regex pattern"
                      />
                    </div>

                    <div>
                      <Label htmlFor="filter_not_title">Title Excludes</Label>
                      <Input
                        id="filter_not_title"
                        value={formData.filter_not_title}
                        onChange={(e) => setFormData({ ...formData, filter_not_title: e.target.value })}
                        placeholder="regex pattern"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="filter_description">Description Contains</Label>
                      <Input
                        id="filter_description"
                        value={formData.filter_description}
                        onChange={(e) => setFormData({ ...formData, filter_description: e.target.value })}
                        placeholder="regex pattern"
                      />
                    </div>

                    <div>
                      <Label htmlFor="filter_not_description">Description Excludes</Label>
                      <Input
                        id="filter_not_description"
                        value={formData.filter_not_description}
                        onChange={(e) => setFormData({ ...formData, filter_not_description: e.target.value })}
                        placeholder="regex pattern"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="filter_min_duration">Min Duration (seconds)</Label>
                      <Input
                        id="filter_min_duration"
                        type="number"
                        value={formData.filter_min_duration}
                        onChange={(e) => setFormData({ ...formData, filter_min_duration: parseInt(e.target.value) || 0 })}
                        placeholder="0"
                      />
                    </div>

                    <div>
                      <Label htmlFor="filter_max_duration">Max Duration (seconds)</Label>
                      <Input
                        id="filter_max_duration"
                        type="number"
                        value={formData.filter_max_duration}
                        onChange={(e) => setFormData({ ...formData, filter_max_duration: parseInt(e.target.value) || 0 })}
                        placeholder="0"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="filter_min_age">Min Age (days)</Label>
                      <Input
                        id="filter_min_age"
                        type="number"
                        value={formData.filter_min_age}
                        onChange={(e) => setFormData({ ...formData, filter_min_age: parseInt(e.target.value) || 0 })}
                        placeholder="0"
                      />
                    </div>

                    <div>
                      <Label htmlFor="filter_max_age">Max Age (days)</Label>
                      <Input
                        id="filter_max_age"
                        type="number"
                        value={formData.filter_max_age}
                        onChange={(e) => setFormData({ ...formData, filter_max_age: parseInt(e.target.value) || 0 })}
                        placeholder="0"
                      />
                    </div>
                  </div>
                </div>
              </TabsContent>

              {/* Metadata Tab */}
              <TabsContent value="metadata">
                <div className="space-y-4">
                  <p className="text-sm text-gray-600 mb-4">
                    Customize podcast metadata. Leave empty to use defaults from the source.
                  </p>

                  <div>
                    <Label htmlFor="custom_title">Podcast Title</Label>
                    <Input
                      id="custom_title"
                      value={formData.custom_title}
                      onChange={(e) => setFormData({ ...formData, custom_title: e.target.value })}
                      placeholder="My Awesome Podcast"
                    />
                  </div>

                  <div>
                    <Label htmlFor="custom_description">Description</Label>
                    <Input
                      id="custom_description"
                      value={formData.custom_description}
                      onChange={(e) => setFormData({ ...formData, custom_description: e.target.value })}
                      placeholder="Podcast description"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="custom_author">Author</Label>
                      <Input
                        id="custom_author"
                        value={formData.custom_author}
                        onChange={(e) => setFormData({ ...formData, custom_author: e.target.value })}
                        placeholder="Author Name"
                      />
                    </div>

                    <div>
                      <Label htmlFor="custom_lang">Language</Label>
                      <Input
                        id="custom_lang"
                        value={formData.custom_lang}
                        onChange={(e) => setFormData({ ...formData, custom_lang: e.target.value })}
                        placeholder="en"
                      />
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="custom_cover_art">Cover Art URL</Label>
                    <Input
                      id="custom_cover_art"
                      value={formData.custom_cover_art}
                      onChange={(e) => setFormData({ ...formData, custom_cover_art: e.target.value })}
                      placeholder="https://example.com/cover.jpg"
                    />
                  </div>

                  <div>
                    <Label htmlFor="custom_cover_art_quality">Cover Art Quality</Label>
                    <Select
                      id="custom_cover_art_quality"
                      value={formData.custom_cover_art_quality}
                      onChange={(e) => setFormData({ ...formData, custom_cover_art_quality: e.target.value })}
                    >
                      <option value="high">High</option>
                      <option value="medium">Medium</option>
                      <option value="low">Low</option>
                    </Select>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="custom_category">Category</Label>
                      <Input
                        id="custom_category"
                        value={formData.custom_category}
                        onChange={(e) => setFormData({ ...formData, custom_category: e.target.value })}
                        placeholder="Technology"
                      />
                    </div>

                    <div>
                      <Label htmlFor="custom_subcategories">Subcategories</Label>
                      <Input
                        id="custom_subcategories"
                        value={formData.custom_subcategories}
                        onChange={(e) => setFormData({ ...formData, custom_subcategories: e.target.value })}
                        placeholder="Tech News, Software"
                      />
                      <p className="text-xs text-gray-500 mt-1">Comma-separated</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="custom_owner_name">Owner Name</Label>
                      <Input
                        id="custom_owner_name"
                        value={formData.custom_owner_name}
                        onChange={(e) => setFormData({ ...formData, custom_owner_name: e.target.value })}
                        placeholder="John Doe"
                      />
                    </div>

                    <div>
                      <Label htmlFor="custom_owner_email">Owner Email</Label>
                      <Input
                        id="custom_owner_email"
                        type="email"
                        value={formData.custom_owner_email}
                        onChange={(e) => setFormData({ ...formData, custom_owner_email: e.target.value })}
                        placeholder="john@example.com"
                      />
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="custom_link">Website Link</Label>
                    <Input
                      id="custom_link"
                      type="url"
                      value={formData.custom_link}
                      onChange={(e) => setFormData({ ...formData, custom_link: e.target.value })}
                      placeholder="https://example.com"
                    />
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="custom_explicit"
                      checked={formData.custom_explicit}
                      onChange={(e) => setFormData({ ...formData, custom_explicit: e.target.checked })}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <Label htmlFor="custom_explicit" className="cursor-pointer">Explicit Content</Label>
                  </div>
                </div>
              </TabsContent>

              {/* Notifications Tab */}
              <TabsContent value="notifications">
                <div className="space-y-4">
                  <p className="text-sm text-gray-600 mb-4">
                    Configure webhook notifications to receive alerts when new episodes are downloaded.
                    The webhook will be called with episode information after each successful download.
                  </p>

                  <div className="flex items-center gap-3 mb-4">
                    <input
                      type="checkbox"
                      id="webhook_enabled"
                      checked={formData.webhook_enabled}
                      onChange={(e) => setFormData({ ...formData, webhook_enabled: e.target.checked })}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <Label htmlFor="webhook_enabled" className="cursor-pointer font-semibold">Enable Webhook Notifications</Label>
                  </div>

                  {formData.webhook_enabled && (
                    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
                      <div>
                        <Label htmlFor="webhook_url">Webhook URL</Label>
                        <Input
                          id="webhook_url"
                          type="url"
                          value={formData.webhook_url}
                          onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
                          placeholder="https://webhook.example.com/notify"
                          required={formData.webhook_enabled}
                        />
                        <p className="text-xs text-gray-500 mt-1">
                          The URL to send HTTP requests to when a new episode is downloaded
                        </p>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <Label htmlFor="webhook_method">HTTP Method</Label>
                          <Select
                            id="webhook_method"
                            value={formData.webhook_method}
                            onChange={(e) => setFormData({ ...formData, webhook_method: e.target.value })}
                          >
                            <option value="POST">POST</option>
                            <option value="GET">GET</option>
                            <option value="PUT">PUT</option>
                            <option value="PATCH">PATCH</option>
                          </Select>
                          <p className="text-xs text-gray-500 mt-1">HTTP method for the webhook request</p>
                        </div>

                        <div>
                          <Label htmlFor="webhook_timeout">Timeout (seconds)</Label>
                          <Input
                            id="webhook_timeout"
                            type="number"
                            min="1"
                            max="300"
                            value={formData.webhook_timeout}
                            onChange={(e) => setFormData({ ...formData, webhook_timeout: parseInt(e.target.value) || 30 })}
                          />
                          <p className="text-xs text-gray-500 mt-1">Maximum wait time for webhook response</p>
                        </div>
                      </div>

                      <div>
                        <Label htmlFor="webhook_message">Message Template</Label>
                        <Input
                          id="webhook_message"
                          value={formData.webhook_message}
                          onChange={(e) => setFormData({ ...formData, webhook_message: e.target.value })}
                          placeholder="New episode: $EPISODE_TITLE"
                        />
                        <p className="text-xs text-gray-500 mt-1">
                          Message sent in the request body. Use variables: $EPISODE_TITLE, $EPISODE_FILE, $FEED_NAME
                        </p>
                      </div>

                      <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                        <p className="text-xs text-blue-800">
                          <strong>Available Variables:</strong>
                          <br/>• <code className="bg-blue-100 px-1 rounded">$EPISODE_TITLE</code> - Title of the downloaded episode
                          <br/>• <code className="bg-blue-100 px-1 rounded">$EPISODE_FILE</code> - Path to the downloaded file
                          <br/>• <code className="bg-blue-100 px-1 rounded">$FEED_NAME</code> - Name of the feed
                          <br/><br/>
                          <strong>Example:</strong> New episode downloaded: $EPISODE_TITLE from $FEED_NAME
                        </p>
                      </div>

                      <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                        <p className="text-xs text-green-800">
                          <strong>How it works:</strong>
                          <br/>After each episode is successfully downloaded, the webhook will be called using curl with your configured settings.
                          The command executed will be similar to:
                          <br/><code className="block mt-2 bg-green-100 p-2 rounded text-xs">
                            curl -X {formData.webhook_method} -d "{formData.webhook_message}" {formData.webhook_url || 'your-webhook-url'}
                          </code>
                        </p>
                      </div>

                      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                        <p className="text-xs text-yellow-800">
                          <strong>Note:</strong> Webhooks are executed asynchronously after episode download completes.
                          If the webhook fails, the episode download is still considered successful.
                          Check your server logs for webhook execution details.
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </TabsContent>

              {/* Advanced Tab */}
              <TabsContent value="advanced">
                <div className="space-y-4">
                  <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-4">
                    <p className="text-sm text-yellow-800">
                      <strong>Warning:</strong> Advanced settings allow fine-tuned control over yt-dlp and post-processing.
                      Only modify these if you understand what you're doing, as incorrect values may cause downloads to fail.
                    </p>
                  </div>

                  <div>
                    <Label htmlFor="youtube_dl_args">Custom yt-dlp Arguments</Label>
                    <Input
                      id="youtube_dl_args"
                      value={formData.youtube_dl_args}
                      onChange={(e) => setFormData({ ...formData, youtube_dl_args: e.target.value })}
                      placeholder='--write-sub, --embed-subs, --sub-lang, en'
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Comma-separated list of extra arguments passed to yt-dlp. Example: --write-sub, --embed-subs, --sub-lang, en
                    </p>
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mt-2">
                      <p className="text-xs text-blue-800">
                        <strong>Examples:</strong>
                        <br/>• Embed English subtitles: <code className="bg-blue-100 px-1 rounded">--write-sub, --embed-subs, --sub-lang, en</code>
                        <br/>• Download thumbnail: <code className="bg-blue-100 px-1 rounded">--write-thumbnail, --embed-thumbnail</code>
                        <br/>• Limit rate: <code className="bg-blue-100 px-1 rounded">--limit-rate, 1M</code>
                        <br/><br/>
                        <strong>Note:</strong> Do not use --audio-format, --format, or --output as they may cause unexpected behavior.
                      </p>
                    </div>
                  </div>

                  <div className="border-t border-gray-200 pt-4 mt-4">
                    <h4 className="text-sm font-semibold text-gray-700 mb-3">Post-Download Script</h4>
                    <p className="text-sm text-gray-600 mb-4">
                      Run a custom script after each episode is downloaded. Useful for transcoding, notifications, or custom processing.
                    </p>

                    <div>
                      <Label htmlFor="post_download_command">Script Command</Label>
                      <Input
                        id="post_download_command"
                        value={formData.post_download_command}
                        onChange={(e) => setFormData({ ...formData, post_download_command: e.target.value })}
                        placeholder="/path/to/your/process-episode.sh"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Absolute path to your post-processing script. The script will receive the episode file path as an argument.
                      </p>
                    </div>

                    <div className="mt-4">
                      <Label htmlFor="post_download_timeout">Script Timeout (seconds)</Label>
                      <Input
                        id="post_download_timeout"
                        type="number"
                        min="1"
                        max="3600"
                        value={formData.post_download_timeout}
                        onChange={(e) => setFormData({ ...formData, post_download_timeout: parseInt(e.target.value) || 120 })}
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Maximum time in seconds for the script to complete (default: 120 seconds)
                      </p>
                    </div>

                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mt-3">
                      <p className="text-xs text-blue-800">
                        <strong>Script Requirements:</strong>
                        <br/>• Must be executable (chmod +x script.sh)
                        <br/>• Should accept the downloaded file path as the first argument
                        <br/>• Exit with code 0 on success, non-zero on failure
                        <br/>• All output (stdout/stderr) will be logged
                      </p>
                    </div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>

            <DialogFooter className="mt-6">
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit">
                {editingFeed ? 'Update Feed' : 'Create Feed'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
};

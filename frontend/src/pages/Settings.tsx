import { useEffect, useState } from 'react';
import { useConfigStore } from '../stores/useConfigStore';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Select } from '../components/ui/select';
import { Label } from '../components/ui/label';
import { Server, Database, Download, HardDrive, Loader2, Save, RefreshCw, Eye, EyeOff, Lock, Clock, Upload } from 'lucide-react';

export const Settings: React.FC = () => {
  const { config, loading, error, loadConfig } = useConfigStore();
  const [serverSettings, setServerSettings] = useState({
    hostname: '',
    port: 8080,
    frontend_port: 5173,
    bind_address: '*',
    tls: false,
    certificate_path: '',
    key_file_path: '',
    path: '',
  });
  const [tlsMode, setTlsMode] = useState<'path' | 'upload'>('path');
  const [certificateFile, setCertificateFile] = useState<File | null>(null);
  const [keyFile, setKeyFile] = useState<File | null>(null);
  const [authSettings, setAuthSettings] = useState({
    enabled: false,
    username: '',
    password: '',
  });
  const [showAuthPassword, setShowAuthPassword] = useState(false);
  const [storageSettings, setStorageSettings] = useState({
    type: 'local',
    data_dir: '',
    s3_bucket: '',
    s3_region: '',
    s3_endpoint_url: '',
    s3_prefix: '',
    s3_access_key: '',
    s3_secret_key: '',
  });
  const [showS3SecretKey, setShowS3SecretKey] = useState(false);
  const [downloaderSettings, setDownloaderSettings] = useState({
    self_update: false,
    update_channel: 'stable',
    update_version: '',
    timeout: '30s',
  });
  const [timeoutMinutes, setTimeoutMinutes] = useState(30);
  const [tokensSettings, setTokensSettings] = useState({
    youtube: '',
    vimeo: '',
    soundcloud: '',
  });
  const [showTokens, setShowTokens] = useState({
    youtube: false,
    vimeo: false,
    soundcloud: false,
  });
  const [historySettings, setHistorySettings] = useState({
    enabled: true,
    retention_days: 30,
    max_entries: 1000,
  });

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  useEffect(() => {
    if (config) {
      setServerSettings({
        hostname: config.server.hostname,
        port: config.server.port,
        frontend_port: config.server.frontend_port || 5173,
        bind_address: config.server.bind_address,
        tls: config.server.tls,
        certificate_path: config.server.certificate_path || '',
        key_file_path: config.server.key_file_path || '',
        path: config.server.path,
      });
      // Set TLS mode based on whether paths are populated
      if (config.server.certificate_path && config.server.key_file_path) {
        setTlsMode('path');
      }
      setAuthSettings({
        enabled: config.server.basic_auth?.enabled || false,
        username: config.server.basic_auth?.username || '',
        password: config.server.basic_auth?.password || '',
      });
      setStorageSettings({
        type: config.storage.type,
        data_dir: config.storage.local?.data_dir || '',
        s3_bucket: config.storage.s3?.bucket || '',
        s3_region: config.storage.s3?.region || '',
        s3_endpoint_url: config.storage.s3?.endpoint_url || '',
        s3_prefix: config.storage.s3?.prefix || '',
        s3_access_key: config.storage.s3?.access_key || '',
        s3_secret_key: config.storage.s3?.secret_key || '',
      });
      setDownloaderSettings({
        self_update: config.downloader.self_update,
        update_channel: config.downloader.update_channel || 'stable',
        update_version: config.downloader.update_version || '',
        timeout: config.downloader.timeout,
      });
      // Parse timeout string (e.g., "30s" or "30m") to minutes
      const timeoutStr = config.downloader.timeout || '30s';
      let minutes = 30;
      if (timeoutStr.endsWith('m')) {
        minutes = parseInt(timeoutStr.slice(0, -1)) || 30;
      } else if (timeoutStr.endsWith('s')) {
        minutes = Math.round((parseInt(timeoutStr.slice(0, -1)) || 30) / 60);
      }
      setTimeoutMinutes(minutes);
      setTokensSettings({
        youtube: config.tokens?.youtube?.join(', ') || '',
        vimeo: config.tokens?.vimeo?.join(', ') || '',
        soundcloud: config.tokens?.soundcloud?.join(', ') || '',
      });
      setHistorySettings({
        enabled: config.history?.enabled ?? true,
        retention_days: config.history?.retention_days ?? 30,
        max_entries: config.history?.max_entries ?? 1000,
      });
    }
  }, [config]);

  const handleSaveServer = async () => {
    try {
      // If TLS is enabled and upload mode is selected, upload certificates first
      if (serverSettings.tls && tlsMode === 'upload' && (certificateFile || keyFile)) {
        const formData = new FormData();
        if (certificateFile) {
          formData.append('certificate', certificateFile);
        }
        if (keyFile) {
          formData.append('key', keyFile);
        }

        const uploadResponse = await fetch('/api/v1/config/tls/upload', {
          method: 'POST',
          body: formData,
        });

        if (!uploadResponse.ok) {
          throw new Error('Failed to upload TLS certificates');
        }

        const uploadData = await uploadResponse.json();
        // Update server settings with the uploaded file paths
        serverSettings.certificate_path = uploadData.certificate_path;
        serverSettings.key_file_path = uploadData.key_file_path;
      }

      const response = await fetch('/api/v1/config/server', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(serverSettings),
      });
      if (!response.ok) throw new Error('Failed to update server settings');
      const data = await response.json();
      alert(data.message || 'Server settings updated successfully! Restart the server for TLS changes to take effect.');
      loadConfig();
    } catch (error) {
      alert('Error updating server settings: ' + (error as Error).message);
    }
  };

  const handleSaveStorage = async () => {
    try {
      const payload: any = {
        type: storageSettings.type,
      };

      if (storageSettings.type === 'local') {
        payload.local = {
          data_dir: storageSettings.data_dir,
        };
      } else if (storageSettings.type === 's3') {
        payload.s3 = {
          bucket: storageSettings.s3_bucket,
          region: storageSettings.s3_region,
          endpoint_url: storageSettings.s3_endpoint_url,
          prefix: storageSettings.s3_prefix,
          access_key: storageSettings.s3_access_key,
          secret_key: storageSettings.s3_secret_key,
        };
      }

      const response = await fetch('/api/v1/config/storage', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error('Failed to update storage settings');
      const data = await response.json();
      alert(data.message || 'Storage settings updated successfully!');
      loadConfig();
    } catch (error) {
      alert('Error updating storage settings: ' + (error as Error).message);
    }
  };

  const handleSaveDownloader = async () => {
    try {
      // Convert minutes to timeout string format (e.g., "30m")
      const timeoutStr = `${timeoutMinutes}m`;
      const response = await fetch('/api/v1/config/downloader', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...downloaderSettings,
          timeout: timeoutStr,
        }),
      });
      if (!response.ok) throw new Error('Failed to update downloader settings');
      const data = await response.json();
      alert(data.message || 'Downloader settings updated successfully!');
      loadConfig();
    } catch (error) {
      alert('Error updating downloader settings: ' + (error as Error).message);
    }
  };

  const handleSaveTokens = async () => {
    try {
      const response = await fetch('/api/v1/config/tokens', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          youtube: tokensSettings.youtube,
          vimeo: tokensSettings.vimeo,
          soundcloud: tokensSettings.soundcloud,
        }),
      });
      if (!response.ok) throw new Error('Failed to update API tokens');
      const data = await response.json();
      alert(data.message || 'API tokens updated successfully!');
      loadConfig();
    } catch (error) {
      alert('Error updating API tokens: ' + (error as Error).message);
    }
  };

  const handleSaveAuth = async () => {
    try {
      const response = await fetch('/api/v1/config/auth', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(authSettings),
      });
      if (!response.ok) throw new Error('Failed to update authentication settings');
      const data = await response.json();
      alert(data.message || 'Authentication settings updated successfully! Restart the server for changes to take effect.');
      loadConfig();
    } catch (error) {
      alert('Error updating authentication settings: ' + (error as Error).message);
    }
  };

  const handleSaveHistory = async () => {
    try {
      const response = await fetch('/api/v1/config/history', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(historySettings),
      });
      if (!response.ok) throw new Error('Failed to update history settings');
      const data = await response.json();
      alert(data.message || 'History settings updated successfully! Restart the server for changes to take effect.');
      loadConfig();
    } catch (error) {
      alert('Error updating history settings: ' + (error as Error).message);
    }
  };

  const handleRestart = async () => {
    if (!confirm('Are you sure you want to restart the server? This will temporarily interrupt service.')) {
      return;
    }

    try {
      const response = await fetch('/api/v1/config/restart', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      if (!response.ok) throw new Error('Failed to restart server');
      const data = await response.json();
      alert(data.message || 'Server restart initiated!');
    } catch (error) {
      alert('Error restarting server: ' + (error as Error).message);
    }
  };

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

  return (
    <div>
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
            <p className="text-gray-600 mt-1">Configure your Podsync server</p>
          </div>
          <Button onClick={handleRestart} variant="destructive" title="Restart the server to apply configuration changes">
            <RefreshCw className="w-4 h-4 mr-2" />
            Restart Server
          </Button>
        </div>
      </div>

      <div className="space-y-6">
        {/* Server Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                <Server className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <CardTitle>Server Configuration</CardTitle>
                <CardDescription>Configure web server and network settings</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <Label htmlFor="hostname">Hostname</Label>
                <Input
                  id="hostname"
                  value={serverSettings.hostname}
                  onChange={(e) => setServerSettings({ ...serverSettings, hostname: e.target.value })}
                  placeholder="http://localhost:8080"
                />
                <p className="text-xs text-gray-500 mt-1">Public URL for RSS feeds</p>
              </div>

              <div>
                <Label htmlFor="port">API Port</Label>
                <Input
                  id="port"
                  type="number"
                  value={serverSettings.port}
                  onChange={(e) => setServerSettings({ ...serverSettings, port: parseInt(e.target.value) })}
                />
                <p className="text-xs text-gray-500 mt-1">Go API server port number</p>
              </div>

              {import.meta.env.DEV && (
                <div>
                  <Label htmlFor="frontend_port">Frontend Port</Label>
                  <Input
                    id="frontend_port"
                    type="number"
                    value={serverSettings.frontend_port}
                    onChange={(e) => setServerSettings({ ...serverSettings, frontend_port: parseInt(e.target.value) })}
                  />
                  <p className="text-xs text-gray-500 mt-1">React dev server port (for development)</p>
                </div>
              )}

              <div>
                <Label htmlFor="bind_address">Bind Address</Label>
                <Input
                  id="bind_address"
                  value={serverSettings.bind_address}
                  onChange={(e) => setServerSettings({ ...serverSettings, bind_address: e.target.value })}
                  placeholder="*"
                />
                <p className="text-xs text-gray-500 mt-1">Network interface (* for all)</p>
              </div>

              <div>
                <Label htmlFor="path">Base Path</Label>
                <Input
                  id="path"
                  value={serverSettings.path}
                  onChange={(e) => setServerSettings({ ...serverSettings, path: e.target.value })}
                  placeholder="/"
                />
                <p className="text-xs text-gray-500 mt-1">URL path prefix</p>
              </div>
            </div>

            <div className="flex items-center gap-3 mb-4">
              <input
                type="checkbox"
                id="tls"
                checked={serverSettings.tls}
                onChange={(e) => setServerSettings({ ...serverSettings, tls: e.target.checked })}
                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
              />
              <Label htmlFor="tls" className="cursor-pointer">Enable TLS/HTTPS</Label>
            </div>

            {serverSettings.tls && (
              <div className="p-4 bg-gray-50 rounded-lg space-y-4 mb-4">
                <div>
                  <Label>Certificate Configuration Method</Label>
                  <div className="mt-2 space-y-2">
                    <div className="flex items-center gap-3">
                      <input
                        type="radio"
                        id="tls-path"
                        name="tls-mode"
                        checked={tlsMode === 'path'}
                        onChange={() => setTlsMode('path')}
                        className="w-4 h-4 text-blue-600 focus:ring-blue-500"
                      />
                      <Label htmlFor="tls-path" className="cursor-pointer">Specify file paths on server</Label>
                    </div>
                    <div className="flex items-center gap-3">
                      <input
                        type="radio"
                        id="tls-upload"
                        name="tls-mode"
                        checked={tlsMode === 'upload'}
                        onChange={() => setTlsMode('upload')}
                        className="w-4 h-4 text-blue-600 focus:ring-blue-500"
                      />
                      <Label htmlFor="tls-upload" className="cursor-pointer">Upload certificate files</Label>
                    </div>
                  </div>
                </div>

                {tlsMode === 'path' && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="certificate_path">Certificate File Path</Label>
                      <Input
                        id="certificate_path"
                        value={serverSettings.certificate_path}
                        onChange={(e) => setServerSettings({ ...serverSettings, certificate_path: e.target.value })}
                        placeholder="/path/to/cert.pem"
                      />
                      <p className="text-xs text-gray-500 mt-1">Path to TLS certificate file on server</p>
                    </div>
                    <div>
                      <Label htmlFor="key_file_path">Private Key File Path</Label>
                      <Input
                        id="key_file_path"
                        value={serverSettings.key_file_path}
                        onChange={(e) => setServerSettings({ ...serverSettings, key_file_path: e.target.value })}
                        placeholder="/path/to/key.pem"
                      />
                      <p className="text-xs text-gray-500 mt-1">Path to TLS private key file on server</p>
                    </div>
                  </div>
                )}

                {tlsMode === 'upload' && (
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="cert-upload">Certificate File (.pem, .crt)</Label>
                      <div className="mt-2 flex items-center gap-3">
                        <Input
                          id="cert-upload"
                          type="file"
                          accept=".pem,.crt,.cer"
                          onChange={(e) => setCertificateFile(e.target.files?.[0] || null)}
                          className="flex-1"
                        />
                        {certificateFile && (
                          <span className="text-sm text-green-600 flex items-center gap-1">
                            <Upload className="w-4 h-4" />
                            {certificateFile.name}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-gray-500 mt-1">Upload your TLS certificate file</p>
                    </div>
                    <div>
                      <Label htmlFor="key-upload">Private Key File (.pem, .key)</Label>
                      <div className="mt-2 flex items-center gap-3">
                        <Input
                          id="key-upload"
                          type="file"
                          accept=".pem,.key"
                          onChange={(e) => setKeyFile(e.target.files?.[0] || null)}
                          className="flex-1"
                        />
                        {keyFile && (
                          <span className="text-sm text-green-600 flex items-center gap-1">
                            <Upload className="w-4 h-4" />
                            {keyFile.name}
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-gray-500 mt-1">Upload your TLS private key file</p>
                    </div>
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                      <p className="text-sm text-blue-800">
                        <strong>Note:</strong> Uploaded certificates will be stored securely on the server.
                        Restart the server after saving for TLS changes to take effect.
                      </p>
                    </div>
                  </div>
                )}

                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                  <p className="text-sm text-yellow-800">
                    <strong>Warning:</strong> Enabling TLS requires valid certificate and key files.
                    The server will fail to start if the files are invalid or cannot be read.
                    Make sure to test your certificates before enabling TLS in production.
                  </p>
                </div>
              </div>
            )}

            <Button onClick={handleSaveServer} title="Apply server configuration changes">
              <Save className="w-4 h-4 mr-2" />
              Save Server Settings
            </Button>
          </CardContent>
        </Card>

        {/* Authentication Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center">
                <Lock className="w-5 h-5 text-red-600" />
              </div>
              <div>
                <CardTitle>Authentication</CardTitle>
                <CardDescription>Password protect your web interface with HTTP Basic Auth</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-3 mb-4">
              <input
                type="checkbox"
                id="auth_enabled"
                checked={authSettings.enabled}
                onChange={(e) => setAuthSettings({ ...authSettings, enabled: e.target.checked })}
                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
              />
              <Label htmlFor="auth_enabled" className="cursor-pointer">Enable HTTP Basic Authentication</Label>
            </div>

            {authSettings.enabled && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4 p-4 bg-gray-50 rounded-lg">
                <div>
                  <Label htmlFor="auth_username">Username</Label>
                  <Input
                    id="auth_username"
                    value={authSettings.username}
                    onChange={(e) => setAuthSettings({ ...authSettings, username: e.target.value })}
                    placeholder="admin"
                  />
                  <p className="text-xs text-gray-500 mt-1">Username for basic auth</p>
                </div>

                <div>
                  <Label htmlFor="auth_password">Password</Label>
                  <div className="relative">
                    <Input
                      id="auth_password"
                      value={authSettings.password}
                      onChange={(e) => setAuthSettings({ ...authSettings, password: e.target.value })}
                      placeholder="Enter secure password"
                      type={showAuthPassword ? "text" : "password"}
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowAuthPassword(!showAuthPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-700 transition-colors"
                      title={showAuthPassword ? "Hide password" : "Show password"}
                    >
                      {showAuthPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">Password for basic auth</p>
                </div>

                <div className="col-span-full bg-blue-50 border border-blue-200 rounded-lg p-3">
                  <p className="text-sm text-blue-800">
                    <strong>Note:</strong> HTTP Basic Authentication will require a username and password
                    when accessing the web interface. This uses standard browser authentication prompts,
                    not a login page. Restart the server after saving for changes to take effect.
                  </p>
                </div>
              </div>
            )}

            <Button onClick={handleSaveAuth} title="Update authentication settings (requires server restart)">
              <Save className="w-4 h-4 mr-2" />
              Save Authentication Settings
            </Button>
          </CardContent>
        </Card>

        {/* Storage Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                <HardDrive className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <CardTitle>Storage Configuration</CardTitle>
                <CardDescription>Configure where media files are stored</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4 mb-4">
              <div>
                <Label htmlFor="storage_type">Storage Type</Label>
                <Select
                  id="storage_type"
                  value={storageSettings.type}
                  onChange={(e) => setStorageSettings({ ...storageSettings, type: e.target.value })}
                >
                  <option value="local">Local Filesystem</option>
                  <option value="s3">Amazon S3</option>
                </Select>
              </div>

              {storageSettings.type === 'local' && (
                <div>
                  <Label htmlFor="data_dir">Data Directory</Label>
                  <Input
                    id="data_dir"
                    value={storageSettings.data_dir}
                    onChange={(e) => setStorageSettings({ ...storageSettings, data_dir: e.target.value })}
                    placeholder="/var/podsync/data"
                  />
                  <p className="text-xs text-gray-500 mt-1">Local directory for media files</p>
                </div>
              )}

              {storageSettings.type === 's3' && (
                <>
                  <div>
                    <Label htmlFor="s3_bucket">S3 Bucket Name</Label>
                    <Input
                      id="s3_bucket"
                      value={storageSettings.s3_bucket}
                      onChange={(e) => setStorageSettings({ ...storageSettings, s3_bucket: e.target.value })}
                      placeholder="my-podsync-bucket"
                    />
                    <p className="text-xs text-gray-500 mt-1">S3 bucket name (required)</p>
                  </div>

                  <div>
                    <Label htmlFor="s3_region">AWS Region</Label>
                    <Input
                      id="s3_region"
                      value={storageSettings.s3_region}
                      onChange={(e) => setStorageSettings({ ...storageSettings, s3_region: e.target.value })}
                      placeholder="us-west-2"
                    />
                    <p className="text-xs text-gray-500 mt-1">AWS region (required)</p>
                  </div>

                  <div>
                    <Label htmlFor="s3_endpoint_url">Endpoint URL</Label>
                    <Input
                      id="s3_endpoint_url"
                      value={storageSettings.s3_endpoint_url}
                      onChange={(e) => setStorageSettings({ ...storageSettings, s3_endpoint_url: e.target.value })}
                      placeholder="https://s3.us-west-2.amazonaws.com"
                    />
                    <p className="text-xs text-gray-500 mt-1">S3 API endpoint (optional, for S3-compatible services)</p>
                  </div>

                  <div>
                    <Label htmlFor="s3_prefix">Bucket Prefix</Label>
                    <Input
                      id="s3_prefix"
                      value={storageSettings.s3_prefix}
                      onChange={(e) => setStorageSettings({ ...storageSettings, s3_prefix: e.target.value })}
                      placeholder="podsync/"
                    />
                    <p className="text-xs text-gray-500 mt-1">Path prefix within bucket (optional)</p>
                  </div>

                  <div>
                    <Label htmlFor="s3_access_key">AWS Access Key ID</Label>
                    <Input
                      id="s3_access_key"
                      value={storageSettings.s3_access_key}
                      onChange={(e) => setStorageSettings({ ...storageSettings, s3_access_key: e.target.value })}
                      placeholder="AKIAIOSFODNN7EXAMPLE"
                    />
                    <p className="text-xs text-gray-500 mt-1">AWS access key for S3 authentication</p>
                  </div>

                  <div>
                    <Label htmlFor="s3_secret_key">AWS Secret Access Key</Label>
                    <div className="relative">
                      <Input
                        id="s3_secret_key"
                        value={storageSettings.s3_secret_key}
                        onChange={(e) => setStorageSettings({ ...storageSettings, s3_secret_key: e.target.value })}
                        placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                        type={showS3SecretKey ? "text" : "password"}
                        className="pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowS3SecretKey(!showS3SecretKey)}
                        className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-700 transition-colors"
                        title={showS3SecretKey ? "Hide secret key" : "Show secret key"}
                      >
                        {showS3SecretKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">AWS secret key for S3 authentication</p>
                  </div>

                  <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                    <p className="text-sm text-yellow-800">
                      <strong>Security Note:</strong> AWS credentials will be stored in the configuration file.
                      For better security in production environments, consider using environment variables
                      (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY) or IAM roles instead.
                    </p>
                  </div>
                </>
              )}
            </div>

            <Button onClick={handleSaveStorage} title="Apply storage configuration changes">
              <Save className="w-4 h-4 mr-2" />
              Save Storage Settings
            </Button>
          </CardContent>
        </Card>

        {/* Downloader Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
                <Download className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <CardTitle>Downloader Configuration</CardTitle>
                <CardDescription>Configure yt-dlp settings and updates</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <Label htmlFor="timeout">Download Timeout (minutes)</Label>
                <Input
                  id="timeout"
                  type="number"
                  value={timeoutMinutes}
                  onChange={(e) => setTimeoutMinutes(parseInt(e.target.value) || 30)}
                  placeholder="30"
                  min="1"
                  max="1440"
                />
                <p className="text-xs text-gray-500 mt-1">Timeout per episode in minutes (e.g., 30 for 30 minutes)</p>
              </div>

              {config?.downloader.ytdl_version && (
                <div>
                  <Label htmlFor="ytdl_version">yt-dlp Version</Label>
                  <Input
                    id="ytdl_version"
                    value={config.downloader.ytdl_version}
                    disabled
                    className="bg-gray-50"
                  />
                  <p className="text-xs text-gray-500 mt-1">Currently installed version</p>
                </div>
              )}
            </div>

            <div className="flex items-center gap-3 mb-4">
              <input
                type="checkbox"
                id="self_update"
                checked={downloaderSettings.self_update}
                onChange={(e) => setDownloaderSettings({ ...downloaderSettings, self_update: e.target.checked })}
                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
              />
              <Label htmlFor="self_update" className="cursor-pointer">Auto-update yt-dlp every 24 hours</Label>
            </div>

            {downloaderSettings.self_update && (
              <div className="p-4 bg-gray-50 rounded-lg space-y-4 mb-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="update_channel">Update Channel</Label>
                    <Select
                      id="update_channel"
                      value={downloaderSettings.update_channel}
                      onChange={(e) => setDownloaderSettings({ ...downloaderSettings, update_channel: e.target.value, update_version: '' })}
                    >
                      <option value="stable">Stable (Recommended)</option>
                      <option value="nightly">Nightly (Daily builds)</option>
                      <option value="master">Master (Latest commits)</option>
                    </Select>
                    <p className="text-xs text-gray-500 mt-1">
                      {downloaderSettings.update_channel === 'stable' && 'Tested releases with stable features'}
                      {downloaderSettings.update_channel === 'nightly' && 'Daily builds with latest patches (recommended for regular users)'}
                      {downloaderSettings.update_channel === 'master' && 'Bleeding edge with latest fixes (may have regressions)'}
                    </p>
                  </div>

                  <div>
                    <Label htmlFor="update_version">Lock to Specific Version (Optional)</Label>
                    <Input
                      id="update_version"
                      value={downloaderSettings.update_version}
                      onChange={(e) => setDownloaderSettings({ ...downloaderSettings, update_version: e.target.value })}
                      placeholder="e.g., 2023.07.06 or stable@2023.07.06"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Leave empty for latest. Format: tag or channel@tag
                    </p>
                  </div>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                  <p className="text-sm text-blue-800">
                    <strong>Update Channels:</strong>
                    <br/>
                    <span className="block mt-1">• <strong>Stable:</strong> Default channel with well-tested changes</span>
                    <span className="block">• <strong>Nightly:</strong> Built daily with latest patches (recommended for active users)</span>
                    <span className="block">• <strong>Master:</strong> Latest commits with newest fixes but may be unstable</span>
                    <span className="block mt-2 text-xs">Note: Restart required after changing update settings</span>
                  </p>
                </div>
              </div>
            )}

            <Button onClick={handleSaveDownloader} title="Update downloader configuration">
              <Save className="w-4 h-4 mr-2" />
              Save Downloader Settings
            </Button>
          </CardContent>
        </Card>

        {/* API Tokens */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
              </div>
              <div>
                <CardTitle>API Tokens</CardTitle>
                <CardDescription>Configure API keys for YouTube, Vimeo, and SoundCloud</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4 mb-4">
              <div>
                <Label htmlFor="youtube_token">YouTube API Key</Label>
                <div className="relative">
                  <Input
                    id="youtube_token"
                    value={tokensSettings.youtube}
                    onChange={(e) => setTokensSettings({ ...tokensSettings, youtube: e.target.value })}
                    placeholder="Enter YouTube API key(s), comma-separated for rotation"
                    type={showTokens.youtube ? "text" : "password"}
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowTokens({ ...showTokens, youtube: !showTokens.youtube })}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-700 transition-colors"
                    title={showTokens.youtube ? "Hide token" : "Show token"}
                  >
                    {showTokens.youtube ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  <a href="https://developers.google.com/youtube/registering_an_application" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                    How to get YouTube API Key
                  </a>
                </p>
              </div>

              <div>
                <Label htmlFor="vimeo_token">Vimeo API Key</Label>
                <div className="relative">
                  <Input
                    id="vimeo_token"
                    value={tokensSettings.vimeo}
                    onChange={(e) => setTokensSettings({ ...tokensSettings, vimeo: e.target.value })}
                    placeholder="Enter Vimeo API key(s), comma-separated for rotation"
                    type={showTokens.vimeo ? "text" : "password"}
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowTokens({ ...showTokens, vimeo: !showTokens.vimeo })}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-700 transition-colors"
                    title={showTokens.vimeo ? "Hide token" : "Show token"}
                  >
                    {showTokens.vimeo ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  <a href="https://developer.vimeo.com/api/guides/start#generate-access-token" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                    How to get Vimeo API Token
                  </a>
                </p>
              </div>

              <div>
                <Label htmlFor="soundcloud_token">SoundCloud API Key</Label>
                <div className="relative">
                  <Input
                    id="soundcloud_token"
                    value={tokensSettings.soundcloud}
                    onChange={(e) => setTokensSettings({ ...tokensSettings, soundcloud: e.target.value })}
                    placeholder="Enter SoundCloud API key(s), comma-separated for rotation"
                    type={showTokens.soundcloud ? "text" : "password"}
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowTokens({ ...showTokens, soundcloud: !showTokens.soundcloud })}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-gray-500 hover:text-gray-700 transition-colors"
                    title={showTokens.soundcloud ? "Hide token" : "Show token"}
                  >
                    {showTokens.soundcloud ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">Optional: For SoundCloud content access</p>
              </div>
            </div>

            <Button onClick={handleSaveTokens} title="Update API keys for YouTube, Vimeo, and SoundCloud">
              <Save className="w-4 h-4 mr-2" />
              Save API Tokens
            </Button>
          </CardContent>
        </Card>

        {/* Database Info */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">
                <Database className="w-5 h-5 text-orange-600" />
              </div>
              <div>
                <CardTitle>Database</CardTitle>
                <CardDescription>Database configuration (read-only)</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-700">
                <span className="font-medium">Directory:</span> {config?.database.dir || 'db'}
              </p>
              <p className="text-xs text-gray-500 mt-2">
                BadgerDB location for metadata and state
              </p>
            </div>
          </CardContent>
        </Card>

        {/* History Settings */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                <Clock className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <CardTitle>History & Job Tracking</CardTitle>
                <CardDescription>Configure history retention and tracking</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="history-enabled">Enable History Tracking</Label>
                  <div className="mt-2 flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="history-enabled"
                      checked={historySettings.enabled}
                      onChange={(e) =>
                        setHistorySettings({ ...historySettings, enabled: e.target.checked })
                      }
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <span className="text-sm font-medium text-gray-900">
                      {historySettings.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    Enable job and operation history tracking
                  </p>
                </div>

                <div>
                  <Label htmlFor="retention-days">Retention Period (days)</Label>
                  <Input
                    id="retention-days"
                    type="number"
                    min="1"
                    max="365"
                    value={historySettings.retention_days}
                    onChange={(e) =>
                      setHistorySettings({
                        ...historySettings,
                        retention_days: parseInt(e.target.value) || 30,
                      })
                    }
                    className="mt-2"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Entries older than this will be cleaned up
                  </p>
                </div>

                <div>
                  <Label htmlFor="max-entries">Maximum Entries</Label>
                  <Input
                    id="max-entries"
                    type="number"
                    min="100"
                    max="100000"
                    value={historySettings.max_entries}
                    onChange={(e) =>
                      setHistorySettings({
                        ...historySettings,
                        max_entries: parseInt(e.target.value) || 1000,
                      })
                    }
                    className="mt-2"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Maximum number of history entries to keep
                  </p>
                </div>
              </div>

              <Button onClick={handleSaveHistory} title="Update history tracking settings (requires server restart)">
                <Save className="w-4 h-4 mr-2" />
                Save History Settings
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

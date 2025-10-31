import { useEffect } from 'react';
import {
  Typography,
  Box,
  Card,
  CardContent,
  CircularProgress,
  Alert,
  TextField,
  Chip,
  Stack,
} from '@mui/material';
import { useConfigStore } from '../stores/useConfigStore';

export const Configuration: React.FC = () => {
  const { config, loading, error, loadConfig } = useConfigStore();

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (!config) {
    return <Alert severity="info">No configuration loaded</Alert>;
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Configuration
      </Typography>

      <Alert severity="info" sx={{ mb: 3 }}>
        Configuration editing is currently view-only. Modify config.toml to make changes.
      </Alert>

      <Stack spacing={3}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Server Settings
            </Typography>
            <Stack spacing={2}>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Hostname"
                  value={config.server.hostname}
                  fullWidth
                  disabled
                />
                <TextField
                  label="Port"
                  value={config.server.port}
                  fullWidth
                  disabled
                />
              </Stack>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Bind Address"
                  value={config.server.bind_address}
                  fullWidth
                  disabled
                />
                <Box display="flex" gap={1} alignItems="center" flex={1}>
                  {config.server.tls && <Chip label="TLS Enabled" color="success" />}
                  {config.server.web_ui && <Chip label="Web UI Enabled" color="primary" />}
                </Box>
              </Stack>
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Storage Settings
            </Typography>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                label="Storage Type"
                value={config.storage.type}
                fullWidth
                disabled
              />
              {config.storage.local && (
                <TextField
                  label="Data Directory"
                  value={config.storage.local.data_dir}
                  fullWidth
                  disabled
                />
              )}
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Configured Feeds
            </Typography>
            <Stack spacing={2}>
              {Object.entries(config.feeds).map(([id, feedConfig]) => (
                <Card variant="outlined" key={id}>
                  <CardContent>
                    <Typography variant="subtitle1" gutterBottom>
                      {id}
                    </Typography>
                    <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                      <TextField
                        label="Quality"
                        value={feedConfig.quality}
                        size="small"
                        fullWidth
                        disabled
                      />
                      <TextField
                        label="Format"
                        value={feedConfig.format}
                        size="small"
                        fullWidth
                        disabled
                      />
                      <TextField
                        label="Update Period"
                        value={feedConfig.update_period}
                        size="small"
                        fullWidth
                        disabled
                      />
                    </Stack>
                  </CardContent>
                </Card>
              ))}
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Downloader Configuration
            </Typography>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField
                label="Timeout"
                value={config.downloader.timeout}
                fullWidth
                disabled
              />
              {config.downloader.ytdl_version && (
                <TextField
                  label="yt-dlp Version"
                  value={config.downloader.ytdl_version}
                  fullWidth
                  disabled
                />
              )}
            </Stack>
            <Box mt={2}>
              {config.downloader.self_update && <Chip label="Self-Update Enabled" color="primary" />}
              {!config.downloader.self_update && <Chip label="Self-Update Disabled" color="default" />}
            </Box>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Database Settings
            </Typography>
            <TextField
              label="Database Directory"
              value={config.database.dir}
              fullWidth
              disabled
            />
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};

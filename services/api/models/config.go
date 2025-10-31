package models

// ConfigResponse represents the application configuration in API responses
type ConfigResponse struct {
	Server     ServerConfig           `json:"server"`
	Storage    StorageConfig          `json:"storage"`
	Feeds      map[string]*FeedConfig `json:"feeds"`
	Database   DatabaseConfig         `json:"database"`
	Downloader DownloaderConfig       `json:"downloader"`
	Tokens     TokensConfig           `json:"tokens"`
}

// TokensConfig represents API tokens configuration
type TokensConfig struct {
	YouTube    []string `json:"youtube,omitempty"`
	Vimeo      []string `json:"vimeo,omitempty"`
	SoundCloud []string `json:"soundcloud,omitempty"`
	Twitch     []string `json:"twitch,omitempty"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Hostname        string `json:"hostname"`
	Port            int    `json:"port"`
	FrontendPort    int    `json:"frontend_port"`
	BindAddress     string `json:"bind_address"`
	TLS             bool   `json:"tls"`
	CertificatePath string `json:"certificate_path,omitempty"`
	KeyFilePath     string `json:"key_file_path,omitempty"`
	Path            string `json:"path"`
	WebUIEnabled    bool   `json:"web_ui"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type  string              `json:"type"`
	Local *LocalStorageConfig `json:"local,omitempty"`
	S3    *S3StorageConfig    `json:"s3,omitempty"`
}

// LocalStorageConfig represents local storage settings
type LocalStorageConfig struct {
	DataDir string `json:"data_dir"`
}

// S3StorageConfig represents S3 storage settings
type S3StorageConfig struct {
	EndpointURL string `json:"endpoint_url"`
	Region      string `json:"region"`
	Bucket      string `json:"bucket"`
	Prefix      string `json:"prefix,omitempty"`
	AccessKey   string `json:"access_key,omitempty"`
	SecretKey   string `json:"secret_key,omitempty"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Dir string `json:"dir"`
}

// DownloaderConfig represents youtube-dl configuration
type DownloaderConfig struct {
	SelfUpdate    bool   `json:"self_update"`
	UpdateChannel string `json:"update_channel,omitempty"`
	UpdateVersion string `json:"update_version,omitempty"`
	Timeout       string `json:"timeout"`
	YtdlVersion   string `json:"ytdl_version,omitempty"`
}

// UpdateConfigRequest represents a request to update configuration
type UpdateConfigRequest struct {
	Server     *ServerConfig          `json:"server,omitempty"`
	Storage    *StorageConfig         `json:"storage,omitempty"`
	Feeds      map[string]*FeedConfig `json:"feeds,omitempty"`
	Database   *DatabaseConfig        `json:"database,omitempty"`
	Downloader *DownloaderConfig      `json:"downloader,omitempty"`
}

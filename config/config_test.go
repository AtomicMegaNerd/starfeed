package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/atomicmeganerd/starfeed/testutils"
)

func TestConfig_NewConfig(t *testing.T) {
	testCases := []struct {
		name            string
		mockCfgData     func() []byte
		expectedConfig  Config
		configLoaderErr error
		expectErr       bool
	}{
		{
			name: "valid config with github forge",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectedConfig: Config{
				Debug:     true,
				SingleRun: true,
				GitForges: []GitForgeConfig{
					{
						Type:  "github",
						Name:  "GitHub",
						Fqdn:  "github.com",
						Token: "ghp_1234567890abcdef",
					},
				},
				RSSServer: RSSServerConfig{
					Name:  "freshrss",
					URL:   "http://freshrss:80",
					User:  "testuser",
					Token: "freshrss_token_12345",
				},
			},
			expectErr: false,
		},
		{
			name: "valid config with forgejo forge",
			mockCfgData: func() []byte {
				return []byte(`
debug = false
single_run = false

[[git_forges]]
type = "forgejo"
name = "Codeberg"
fqdn = "codeberg.org"
token = "forgejo_token_123456"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectedConfig: Config{
				Debug:     false,
				SingleRun: false,
				GitForges: []GitForgeConfig{
					{
						Type:  "forgejo",
						Name:  "Codeberg",
						Fqdn:  "codeberg.org",
						Token: "forgejo_token_123456",
					},
				},
				RSSServer: RSSServerConfig{
					Name:  "freshrss",
					URL:   "http://freshrss:80",
					User:  "testuser",
					Token: "freshrss_token_12345",
				},
			},
			expectErr: false,
		},
		{
			name: "valid config with multiple forges",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = false

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[[git_forges]]
type = "forgejo"
name = "Codeberg"
fqdn = "codeberg.org"
token = "forgejo_token_123456"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectedConfig: Config{
				Debug:     true,
				SingleRun: false,
				GitForges: []GitForgeConfig{
					{
						Type:  "github",
						Name:  "GitHub",
						Fqdn:  "github.com",
						Token: "ghp_1234567890abcdef",
					},
					{
						Type:  "forgejo",
						Name:  "Codeberg",
						Fqdn:  "codeberg.org",
						Token: "forgejo_token_123456",
					},
				},
				RSSServer: RSSServerConfig{
					Name:  "freshrss",
					URL:   "http://freshrss:80",
					User:  "testuser",
					Token: "freshrss_token_12345",
				},
			},
			expectErr: false,
		},
		{
			name: "missing git_forges",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "empty git_forges array",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

git_forges = []

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "invalid forge type",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "gitlab"
name = "GitLab"
fqdn = "gitlab.com"
token = "gitlab_token_123456"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "forge name too short",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GH"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "forge fqdn too short",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "gh.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "forge token too short",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "short"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "invalid rss server name",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "miniflux"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "invalid rss server url",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "not-a-url"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "rss server user too short",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "ab"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "rss server token too short",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "short"
`)
			},
			expectErr: true,
		},
		{
			name: "unknown toml field",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true
unknown_field = "value"

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "ghp_1234567890abcdef"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "testuser"
token = "freshrss_token_12345"
`)
			},
			expectErr: true,
		},
		{
			name: "invalid toml syntax",
			mockCfgData: func() []byte {
				return []byte(`
debug = true
single_run = true
[invalid toml
`)
			},
			expectErr: true,
		},
		{
			name: "config loader error",
			mockCfgData: func() []byte {
				return []byte("")
			},
			configLoaderErr: errors.New("failed to load config"),
			expectErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockCfgLoader := testutils.MockConfigLoader{
				ExpectedData:  tc.mockCfgData(),
				ExpectedError: tc.configLoaderErr,
			}

			cfg, err := NewConfig(mockCfgLoader)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got %v", err)
			}

			if !reflect.DeepEqual(tc.expectedConfig, cfg) {
				t.Fatalf("Expected %v but got %v", tc.expectedConfig, cfg)
			}
		})
	}
}

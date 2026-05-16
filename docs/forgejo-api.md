# Forgejo API Reference for Starfeed

Forgejo (which powers Codeberg) provides a Gitea-compatible REST API. This document covers the
endpoints and patterns needed for the starfeed integration.

## Base URL

All API requests are made to:

```text
{baseURL}/api/v1/
```

Examples:

- Codeberg: `https://codeberg.org/api/v1/`
- Self-hosted: `https://forgejo.example.com/api/v1/`

## API Reference

The Swagger-generated API docs are available at: `{baseURL}/api/swagger` The OpenAPI spec is at:
`{baseURL}/swagger.v1.json`

## 1. List Starred Repos

### Request

```text
GET {baseURL}/api/v1/user/starred
```

**Authentication:** Required. Forgejo supports `Bearer`, `token`, and Basic Auth:

```text
Authorization: Bearer {api_token}
```

### Query Parameters

| Parameter | Type | Default | Max | Description    |
| --------- | ---- | ------- | --- | -------------- |
| `limit`   | int  | 30      | 50  | Items per page |
| `page`    | int  | 1       | -   | Page number    |

### Pagination

Uses the `Link` response header, identical to GitHub:

```text
link: <https://codeberg.org/api/v1/user/starred?limit=50&page=2>; rel="next",
      <https://codeberg.org/api/v1/user/starred?limit=50&page=10>; rel="last"
```

Regex to extract next page URL:

```go
nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
```

The `x-total-count` header provides the total number of items.

### Response

Array of `Repository` objects. Key fields:

```json
{
  "id": 1092437,
  "owner": {
    "login": "just-shadyumbrella",
    "html_url": "https://codeberg.org/just-shadyumbrella"
  },
  "name": "_",
  "full_name": "just-shadyumbrella/_",
  "html_url": "https://codeberg.org/just-shadyumbrella/_",
  "url": "https://codeberg.org/api/v1/repos/just-shadyumbrella/_",
  "stars_count": 0,
  "release_counter": 0,
  "has_releases": false,
  "archived": false,
  "private": false,
  "created_at": "2025-12-31T03:32:08+01:00",
  "updated_at": "2026-02-26T22:37:53+01:00"
}
```

## 2. Release Atom Feeds

Feed URL pattern (identical to GitHub):

```text
{html_url}/releases.atom
```

Example: `https://codeberg.org/forgejo/forgejo/releases.atom`

Regex to match a Forgejo release feed:

```go
isRelRepoRegex := regexp.MustCompile(
    fmt.Sprintf(`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`, regexp.QuoteMeta(baseURL)),
)
```

## 3. API Settings

Get pagination defaults and limits:

```text
GET {baseURL}/api/v1/settings/api
```

Response:

```json
{
  "max_response_items": 50,
  "default_paging_num": 30,
  "default_git_trees_per_page": 1000,
  "default_max_blob_size": 10485760
}
```

## Key Differences from GitHub

| Aspect            | GitHub                          | Forgejo/Codeberg                |
| ----------------- | ------------------------------- | ------------------------------- |
| Auth header       | `Authorization: Bearer {token}` | `Authorization: Bearer {token}` |
| Pagination param  | `per_page`                      | `limit`                         |
| Default page size | 30                              | 30                              |
| Max page size     | 100                             | 50                              |
| Endpoint prefix   | `/user/starred`                 | `/api/v1/user/starred`          |
| Feed URL          | `{html_url}/releases.atom`      | `{html_url}/releases.atom`      |

## Authentication

Forgejo supports:

- `Authorization: Bearer {api_token}`
- `Authorization: token {api_token}`
- HTTP Basic Auth

For two-factor auth, add the `X-Forgejo-OTP: {code}` header.

# Starfeed: GitHub Subscribed Issues & Pull Requests Feeds

## GitHub API Endpoints

### 1. List Subscribed Issues and Pull Requests

```text
GET /issues?filter=subscribed&state=all&per_page=100
```

**Purpose:** Returns all issues *and* pull requests that the authenticated user has explicitly
subscribed to. GitHub models PRs as issues, so both types come from this single endpoint.

#### Key query parameters

| Parameter  | Value        | Notes                                         |
|------------|--------------|-----------------------------------------------|
| `filter`   | `subscribed` | Only explicitly subscribed items              |
| `state`    | `all`        | Include open and closed items                 |
| `per_page` | `100`        | Maximum allowed; use pagination for more      |
| `since`    | ISO 8601     | Optional: only items updated after this time  |

**Pagination:** Follow the `Link` header with `rel="next"` — identical to the existing starred
repos pagination in `github/github.go`.

**Distinguishing issues from PRs:** Check for the presence of the `pull_request` field. It is
absent on issues and present on PRs:

#### Issue — no pull_request field

```json
{
  "id": 123456789,
  "number": 1234,
  "title": "Add generics support",
  "body": "Issue description...",
  "html_url": "https://github.com/golang/go/issues/1234",
  "state": "open",
  "updated_at": "2026-04-15T10:00:00Z",
  "created_at": "2026-01-01T00:00:00Z",
  "user": { "login": "gopher" },
  "repository_url": "https://api.github.com/repos/golang/go",
  "labels": [{ "name": "enhancement" }]
}
```

#### Pull Request — has pull_request field

```json
{
  "id": 234567890,
  "number": 5678,
  "title": "feat: add worker pool for goroutines",
  "body": "PR description...",
  "html_url": "https://github.com/golang/go/pull/5678",
  "state": "open",
  "updated_at": "2026-04-14T09:00:00Z",
  "created_at": "2026-04-10T00:00:00Z",
  "user": { "login": "gopher" },
  "repository_url": "https://api.github.com/repos/golang/go",
  "pull_request": {
    "url": "https://api.github.com/repos/golang/go/pulls/5678",
    "html_url": "https://github.com/golang/go/pull/5678"
  }
}
```

**Derive the repo:** Parse `repository_url` to extract `owner` and `repo`:

```text
https://api.github.com/repos/{owner}/{repo}
```

**Auth:** Bearer token — same as the existing GitHub client.

### 2. List Comments for an Issue or PR (Conversation Comments)

```text
GET /repos/{owner}/{repo}/issues/{number}/comments?per_page=100
```

**Purpose:** Fetch general conversation comments on either an issue or a PR. This endpoint works
for both — PRs share the issues comment thread. Call once per subscribed item per run.

**Key query parameters:**

| Parameter  | Value    | Notes                                           |
|------------|----------|-------------------------------------------------|
| `per_page` | `100`    | Maximum allowed; paginate if thread is busy     |
| `since`    | ISO 8601 | Optional: only comments created after this time |

**Key response fields per comment:**

```json
{
  "id": 987654321,
  "html_url": "https://github.com/golang/go/issues/1234#issuecomment-987654321",
  "body": "Comment text...",
  "user": { "login": "rsc" },
  "created_at": "2026-02-01T12:00:00Z",
  "updated_at": "2026-02-01T12:00:00Z"
}
```

Note: `html_url` uses `#issuecomment-{id}` for both issue comments and PR conversation comments.

**Include the file path and line in the entry title** so the feed item is immediately useful:

```text
golang/go #5678: feat: add worker pool (review by bradfitz on src/cmd/go/main.go:42)
```

---

### 4. List PR Reviews (Approve / Request Changes / Comment)

```text
GET /repos/{owner}/{repo}/pulls/{number}/reviews?per_page=100
```

**Purpose:** Fetch top-level review submissions — the approve/request-changes/comment actions
that wrap inline comments. These are meaningful feed events ("bradfitz approved", "rsc requested
changes").

**Key response fields per review:**

```json
{
  "id": 556677889,
  "html_url": "https://github.com/golang/go/pull/5678#pullrequestreview-556677889",
  "body": "LGTM, one small nit below.",
  "state": "APPROVED",
  "user": { "login": "bradfitz" },
  "submitted_at": "2026-04-12T08:05:00Z"
}
```

`state` values: `APPROVED`, `CHANGES_REQUESTED`, `COMMENTED`, `DISMISSED`.

Use `html_url` as the Atom `<id>`. Include `state` in the entry title:

```text
golang/go #5678: feat: add worker pool (review APPROVED by bradfitz)
```

**Rate limit consideration:** With N subscribed items (issues + PRs), the total API calls per
daily run are:

- 1 call for the combined issues/PRs list
- N calls for issue/PR conversation comments
- P calls for PR review comments (PRs only)
- P calls for PR reviews (PRs only)

Where P ≤ N. This is well within the 5,000 requests/hour authenticated rate limit for typical
use.

## Synthetic Atom Feed Design

The app generates two feed types per repository: one for subscribed issues, one for subscribed
PRs. Each feed is served over HTTP and registered with FreshRSS.

### Feed URL structure

```text
http://{starfeed-host}:{port}/github/{owner}/{repo}/issues.atom
http://{starfeed-host}:{port}/github/{owner}/{repo}/pr.atom
```

Examples:

```text
http://starfeed:8080/github/golang/go/issues.atom
http://starfeed:8080/github/golang/go/pr.atom
```

Note: `.atom` is used for both for consistency, though `.xml` is equally valid since Atom is XML.
Adjust to `.xml` for the PR feed if preferred — just be consistent.

### Atom entry IDs

Use the GitHub HTML URL as the Atom `<id>` for every entry type. These URLs are permanent and
globally unique. FreshRSS deduplicates by `<id>`, so no app-side state is needed.

| Entry type          | Atom `<id>` value                                            |
|---------------------|--------------------------------------------------------------|
| Issue itself        | `html_url` — e.g. `.../issues/1234`                          |
| Issue comment       | `html_url` — e.g. `...#issuecomment-987654321`               |
| PR itself           | `html_url` — e.g. `.../pull/5678`                            |
| PR conv. comment    | `html_url` — e.g. `...#issuecomment-987654321`               |
| PR review comment   | `html_url` — e.g. `...#discussion_r112233445`                |
| PR review           | `html_url` — e.g. `...#pullrequestreview-556677889`          |

### Atom entry examples

#### Issue itself

```xml
<entry>
  <id>https://github.com/golang/go/issues/1234</id>
  <title>golang/go #1234: Add generics support</title>
  <link href="https://github.com/golang/go/issues/1234"/>
  <updated>2026-01-01T00:00:00Z</updated>
  <author><name>gopher</name></author>
  <content type="text">Issue description...</content>
</entry>
```

#### Issue or PR conversation comment

```xml
<entry>
  <id>https://github.com/golang/go/issues/1234#issuecomment-987654321</id>
  <title>golang/go #1234: Add generics support (comment by rsc)</title>
  <link href="https://github.com/golang/go/issues/1234#issuecomment-987654321"/>
  <updated>2026-02-01T12:00:00Z</updated>
  <author><name>rsc</name></author>
  <content type="text">Comment text...</content>
</entry>
```

#### PR review (approve/request changes)

```xml
<entry>
  <id>https://github.com/golang/go/pull/5678#pullrequestreview-556677889</id>
  <title>golang/go #5678: feat: add worker pool (review APPROVED by bradfitz)</title>
  <link href="https://github.com/golang/go/pull/5678#pullrequestreview-556677889"/>
  <updated>2026-04-12T08:05:00Z</updated>
  <author><name>bradfitz</name></author>
  <content type="text">LGTM, one small nit below.</content>
</entry>
```

#### PR inline review comment

```xml
<entry>
  <id>https://github.com/golang/go/pull/5678#discussion_r112233445</id>
  <title>golang/go #5678: feat: add worker pool (review by bradfitz on
    src/cmd/go/main.go:42)</title>
  <link href="https://github.com/golang/go/pull/5678#discussion_r112233445"/>
  <updated>2026-04-12T08:00:00Z</updated>
  <author><name>bradfitz</name></author>
  <content type="text">Nit: consider renaming this variable.</content>
</entry>
```

Include the issue/PR itself as an entry so the original post appears alongside its comments.

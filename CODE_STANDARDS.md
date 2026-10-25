# Code Standards

## Style
- Minimal code: no comments, no magic numbers (name constants), no redundant abstractions.
- No tests unless they add real value; skips are fine to remove.
- Small, single-purpose PRs with verification output in the PR body.

## Commits
- No PII (IPs, hostnames, emails, tokens) in commits, PR bodies, or outputs.
- Author commits as R's identity (Roeeeee/Roe3, noreply email) — never machine defaults like "Super User <root@ip-...>".
- If a commit leaks PII, squash-merge or force-push before it reaches a public repo.

## Infra / Config
- Keep live systemd/config state in sync with repo definitions.
- Schedule cron jobs in Israel local time (UTC+3), not server UTC.

<!--
SHOW HN

Submit at: https://news.ycombinator.com/submit
Title: Show HN: Dronagiri – a terminal package manager for macOS
URL: https://mastershivnath007.github.io/dronagiri/

Post the first comment immediately after submitting.
Do not ask for upvotes. Do not ask friends to upvote. Reply to comments.
-->

## Title

```
Show HN: Dronagiri – a terminal package manager for macOS
```

## URL

```
https://mastershivnath007.github.io/dronagiri/
```

## First comment

```
Author here. I built this because brew list gives me 188 names and no context
about which 145 are dependencies I never chose. Think Cydia/Sileo, but for the
packages on a Mac.

Dronagiri reads Homebrew, npm, pipx, Cargo, RubyGems and go install through
their documented commands, measures install paths with du, and asks brew deps
for real runtime linkage instead of trusting the declared dependency list. The
UI is Go + Bubble Tea. Removal is delegated to the package manager and requires
two confirmations; there is no --force.

Things it deliberately does not claim: last used is only as good as Spotlight
metadata or timestamped shell history, so it says "never" rather than inventing
a date. Popularity is Homebrew's public analytics, so non-brew packages show no
stars.

macOS 13+, MIT, no telemetry:
https://github.com/MasterShivnath007/dronagiri

Happy to answer questions about the dependency graph or the TUI.
```

## Timing

Post on a weekday morning in the US, roughly:

- 8:00–10:00 AM Eastern
- 5:30–7:30 PM India

Tuesday to Thursday is usually better than Friday or the weekend. HN moves
quickly; stay available for the first two to three hours.

## Rules

- Use the website URL, not the dev.to article or the GitHub repo.
- Do not add “please upvote”, “feedback welcome” or marketing language to the title.
- Do not ask friends or communities to upvote. HN detects voting rings.
- Reply to substantive criticism with facts, not defensiveness.
- If it does not get traction, do not repost the same URL immediately. Wait a
  few weeks and try a different angle.

## Anticipated questions

**“This is not a package manager.”**
It does not install packages. It manages the inventory across package managers:
sizes, dates, dependencies, broken entries, updates and removal. The title is
the shortest accurate description; the first comment says what it does not do.

**“Why not just use brew leaves / brew autoremove?”**
Those are good for Homebrew only. Dronagiri also reads npm, pipx, Cargo,
RubyGems and go install, and adds sizes, install dates, last-used evidence and a
combined view. The guides explain the Homebrew-only commands if that is all
someone needs.

**“Last used is unreliable.”**
Agreed. It uses Spotlight metadata for .app bundles and timestamped shell
history for commands. When neither exists it prints “never”, which means “no
trace”, not “never ran”. The detail panel says so.

**“How do you know the dependencies?”**
brew deps gives the real runtime linkage; the declared JSON dependencies are
often wrong. The graph is inverted so each library can name its dependents.

**“Does it need sudo?”**
No for scanning or listing. Some cask removals ask for a password because macOS
owns those files.

**“Why Go and Bubble Tea?”**
Single binary, fast startup, easy TUI testing, and a small dependency surface.

**“Is it safe?”**
Removal is delegated to the package manager, requires two confirmations, skips
macOS-owned packages, and never adds --force or --ignore-dependencies.

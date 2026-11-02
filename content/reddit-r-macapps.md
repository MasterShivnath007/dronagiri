<!--
REDDIT — r/macapps

Post type: text post. If the editor lets you attach an image, add
docs/screenshot.png as well.

Flair: choose the "App" / "Self-Promotion" flair that r/macapps requires.
Read the sidebar before posting. If the post is removed, read the mod message;
do not repost the same text.

Suggested time: weekday, 9:00–11:00 AM Eastern / 6:30–8:30 PM India.
-->

## Title

```
I built a free macOS package manager that puts brew, npm, pipx, Cargo, gems and Go in one list
```

## Body

```
I kept losing track of what was actually installed on my Mac. brew list showed
188 entries, and most of them were dependencies I never chose. npm, pipx and
Cargo each had their own separate list.

Dronagiri is a terminal app that combines them:

- real on-disk size, install date and last-used evidence
- a Mine / Dependencies split, so libraries are not mixed in with the tools you chose
- tabs for updates and broken entries
- two confirmations before removing anything
- removal delegated to the package manager that installed it

It is not a GUI app. It runs in Terminal, works over SSH, and is for macOS 13+.

Screenshot:
https://raw.githubusercontent.com/MasterShivnath007/dronagiri/main/docs/screenshot.png

Website:
https://mastershivnath007.github.io/dronagiri/

Source:
https://github.com/MasterShivnath007/dronagiri

It is free and MIT-licensed. I am the author, so I am biased, but I would like
to hear what breaks on your machine — especially if you use a package manager I
do not cover.
```

## If it gets removed

- Check the mod message first. Many subreddits require a specific flair, a
  minimum account age, or a self-promotion day.
- Do not repost immediately. Message the moderators once, politely, and ask what
  to change.
- Do not cross-post the same text to r/commandline and r/golang on the same day.

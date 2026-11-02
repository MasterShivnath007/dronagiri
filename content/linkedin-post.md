<!--
LINKEDIN POST

Image to attach:
docs/og.png (1200 x 630)

Optional first comment:
The full build story is on dev.to: https://dev.to/master_shivnath_007/i-built-a-cli-package-manager-for-macos-because-brew-list-hid-145-dependencies-19e5

Do not paste this comment into the post.
-->

brew list on my Mac showed 188 packages.

About 145 were dependencies I never chose. They arrived because something else
needed them. I had no easy way to see which was which, how big they were, or what
was safe to remove.

So I built Dronagiri: a free, open-source terminal package manager for macOS.

It reads Homebrew, npm, pipx, Cargo, RubyGems and Go packages and puts them in
one sortable table:

→ real disk size
→ install date
→ last-used evidence
→ dependencies vs. the tools I actually chose
→ broken entries
→ two-step safe removal

The part I use most is the Mine/Dependencies split. brew list treats the 40
tools I chose and the 145 libraries they pulled in as one flat list. Dronagiri
separates them.

It is macOS 13+, MIT-licensed, and there is no telemetry.

Install:
go install github.com/MasterShivnath007/dronagiri@latest

Website and install guide:
https://mastershivnath007.github.io/dronagiri/

If you try it, I would like to know what your Mine/Dependencies split looks like
— and which package you were surprised to find.

#macOS #OpenSource #DeveloperTools #Homebrew

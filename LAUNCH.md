# Launch and discoverability checklist

This file is for maintainers. It records what is already in place and what still
needs a human account (Google, social, directories).

## Already in place

- README with keyword-relevant headings: "macOS package manager",
  "Homebrew uninstaller", "`brew list` alternative", "Cydia for macOS".
- A real screenshot instead of a text-only table.
- GitHub Pages landing page with:
  - unique title and meta description
  - canonical URL
  - Open Graph and Twitter card images (`docs/og.png`)
  - `SoftwareApplication` and `FAQPage` structured data
  - `sitemap.xml` and `robots.txt`
  - fast, dependency-free HTML/CSS (no external fonts or scripts)
- Repository description, topics and homepage link.
- Release binaries for Apple Silicon and Intel.

## 1. Google Search Console (10 minutes, one time)

Google will not know about a new GitHub Pages site until it crawls a link or a
sitemap. Do this once the Pages URL is live:

1. Open <https://search.google.com/search-console>.
2. Add a **URL prefix** property:
   `https://mastershivnath007.github.io/dronagiri/`
3. Verify with the **HTML tag** method. Copy the `<meta name="google-site-verification" ...>`
   tag into the `<head>` of `docs/index.html`, push, then click Verify.
   (The HTML file method also works: put the file Google gives you in `docs/`.)
4. Open **Sitemaps** and submit:
   `https://mastershivnath007.github.io/dronagiri/sitemap.xml`
5. Use **URL Inspection → Request indexing** for the homepage.
6. Repeat for Bing Webmaster Tools (<https://www.bing.com/webmasters>) and
   import the property from Search Console.

Expect days to weeks, not minutes. A new repository with no backlinks will not
outrank established pages immediately, regardless of on-page work.

## 2. Backlinks that actually help

Search engines rank pages partly by who links to them. A handful of real links
beat hundreds of directory listings.

- Add the site to your GitHub profile README and personal site.
- Post a short write-up on dev.to, Hashnode, Medium or LinkedIn. Link to the
  Pages site, not only the repository.
- Answer relevant questions on Stack Overflow / Reddit with a link only when
  the tool genuinely solves the problem.
- Submit to curated lists (open a PR, do not spam):
  - `awesome-mac`
  - `awesome-cli-apps`
  - `awesome-macos`
  - `awesome-homebrew`
- Submit to Product Hunt and Hacker News **Show HN** once, with a real story:
  "I built a Cydia-style package manager for macOS because brew list hid 145
  dependencies".
- Ask a few Mac developer newsletters or blogs for a mention. Personalised
  emails work; mass mail does not.

## 3. Distribution that creates search demand

- Add a Homebrew tap (`brew install MasterShivnath007/dronagiri/dronagiri`) so
  Homebrew users can install without Go.
- Publish a signed and notarized binary to remove the Gatekeeper warning.
- Keep the release notes keyword-relevant and link to the Pages site.
- Add a screenshot to the GitHub social preview (Settings → General → Social
  preview). The API cannot set this; upload `docs/og.png` manually.

## 4. Content ideas

Each of these is a real search query:

- "How to see what Homebrew packages are taking up space on macOS"
- "How to uninstall npm global packages on macOS"
- "brew list shows 188 packages — which ones are dependencies?"
- "How to find unused CLI tools on a Mac"
- "Cydia for macOS: one package manager for brew, npm, pipx and Cargo"
- "How to clean up a Mac developer environment without deleting the wrong thing"

Publish them as posts that link to the site and the repository. Do not copy the
same text to every platform; search engines discount duplicates.

## 5. What not to do

- Do not buy backlinks or use link farms.
- Do not stuff keywords into the README; it reads badly and does not help.
- Do not create dozens of near-duplicate pages for every package manager.
- Do not promise "top of Google". Ranking is earned over time.

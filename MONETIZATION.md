# Monetization plan

A candid plan for turning Dronagiri into income without damaging the open-source
project.

## The honest starting point

Open-source donations are slow. A new repository with zero stars will not earn
rent from sponsorships in the first month. The realistic order is:

1. **Paid services** — someone pays for your time and judgement.
2. **Sponsorship** — recurring, small, grows with visibility.
3. **Products** — a Pro build or a team tool, only after there is demand.

Do not paywall the MIT code that already exists. You cannot revoke MIT for
released versions, and a hostile license change would kill adoption. Sell
convenience, support and custom work instead.

## What is already live

- `SUPPORT.md` with sponsorship tiers, a $49 Mac audit, team onboarding,
  custom-feature sponsorship and commercial support.
- The website support section: <https://mastershivnath007.github.io/dronagiri/#support>
- `.github/FUNDING.yml` pointing at GitHub Sponsors once the profile is active.
- A public issue link for audit requests, so no backend is required.

## Week 1 — switch on payments

- [ ] Enable GitHub Sponsors for `MasterShivnath007`
      (<https://github.com/sponsors/accounts>). India is supported; payouts go
      through Stripe to an Indian bank account. Add tiers:
      - $5/month — Supporter
      - $25/month — Backer (priority triage)
      - $100/month — Company (credit + priority fixes)
- [ ] Create a Ko-fi page for one-off international cards. Add the link to
      `SUPPORT.md` and the site.
- [ ] For Indian supporters, use GitHub Sponsors. Do not post a UPI ID or bank
      account number. Razorpay is paused until its public page is sanitised.
- [ ] Add a "Hire me" line to the GitHub profile and LinkedIn:
      "I build small, careful macOS developer tools. Dronagiri is one of them."
- [ ] Set the GitHub profile location/website so people can find the project.

## Week 2 — launch and collect proof

- [ ] Show HN: "I built a Cydia-style package manager for macOS because brew
      list hid 145 dependencies". Link the website, not only the repo.
- [ ] Post to r/macapps, r/commandline, r/golang and r/homebrew. Follow each
      community's self-promotion rules; answer questions, do not drop links.
- [ ] Product Hunt launch with the OG image and a 30-second demo GIF.
- [ ] Publish two articles:
      - "What 145 Homebrew dependencies actually are"
      - "How to clean up a Mac developer environment without deleting the wrong thing"
- [ ] Ask five Mac developers to try it and give one sentence of feedback. Turn
      the best quotes into a testimonial block (with permission).

## Week 3 — sell the audit

The $49 audit is the fastest path to the first payment.

- [ ] Create a Fiverr gig: "I will audit and clean up your Mac developer
      environment". Use the copy below.
- [ ] Create an Upwork profile with the same offer.
- [ ] Post on LinkedIn with a before/after size report and a clear call to
      action.
- [ ] Send 20 personalised messages to developers who complain about disk space
      or Homebrew clutter. No mass mail.
- [ ] Offer the first three audits at $29 in exchange for a testimonial, then
      return to $49.

### Fiverr gig copy

**Title:** I will audit and clean up your Mac developer environment

**Description:**

> Your Mac is full of tools you installed once and forgot. I will use Dronagiri,
> an open-source package manager I built, to inventory Homebrew, npm, pipx,
> Cargo, RubyGems and Go packages, find the real disk hogs, separate your tools
> from dependencies, and give you a safe removal plan.
>
> You get:
> - a full inventory with sizes and install dates
> - a prioritised keep/remove list
> - broken entries and unused dependencies called out
> - a 30-minute call to walk through it
> - seven days of follow-up questions
>
> I will not delete anything without your approval. macOS-owned packages are
> marked protected.

**Packages:** Basic $29 (inventory + report), Standard $49 (report + 30-minute
call), Premium $99 (report + call + shell-history/Spotlight setup + follow-up).

**Tags:** mac cleanup, homebrew, disk space, developer tools, macos.

## Week 4 — compound

- [ ] Add the first testimonials to `SUPPORT.md` and the site.
- [ ] Raise the audit to $79 if demand appears.
- [ ] Ask every happy user for a GitHub star and a one-line review.
- [ ] Track: website visits, GitHub stars, release downloads, sponsor count,
      audit enquiries, paid audits. One spreadsheet is enough.
- [ ] Kill what does not work. Double down on the channel that produced a paid
      conversation.

## Product ideas, in order of effort

1. **Signed and notarized binary** — reduces install friction and is a credible
   paid convenience. Apple Developer membership is $99/year; do not buy it until
   revenue covers it.
2. **Dronagiri Pro** — a separate paid build with scheduled inventory reports,
   auto-update checks and a team dashboard. Keep the core MIT; the Pro add-on can
   be closed source or dual-licensed. Build it only after 10+ people ask.
3. **Team onboarding package** — $149–$499 per team, one session plus a runbook.
   This is consulting, not software, and can start today.
4. **Homebrew tap / one-line install** — distribution, not revenue, but it
   increases the top of the funnel.
5. **Sponsor-a-feature** — $250+ for a specific adapter or integration. The code
   stays MIT; the sponsor gets priority and credit.

## What not to do

- Do not switch the license to "non-commercial" or "source-available" for the
  existing code. It violates the MIT grant and destroys trust.
- Do not add telemetry or ads to a tool that markets privacy.
- Do not promise support you cannot deliver. A refund is cheaper than an angry
  customer.
- Do not buy stars, reviews or backlinks.
- Do not quit your income plan on the hope of donations.

## First-month target

A realistic first month is **one paid audit ($29–$49)** and **5–20 GitHub
stars**, not $1,000 from sponsors. The point is proof: a testimonial, a case
study and a repeatable offer. Revenue compounds from there.

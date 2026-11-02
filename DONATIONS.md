# Donation and sponsorship setup

Two rails, deliberately:

1. **GitHub Sponsors** — recurring, international, built into the repository.
2. **Razorpay Payment Page** — one-time Indian payments (UPI, cards, netbanking)
   and the paid Mac audit.

No Ko-fi, no public UPI QR.

## Why these two

| | GitHub Sponsors | Razorpay |
|---|---|---|
| Recurring | Yes | No (one-time) |
| International | Yes | Possible, but GitHub is better for it |
| Indian UPI/cards | No | Yes |
| Where it lives | GitHub repo Sponsor button | A hosted payment page |
| Payout | Stripe → Indian bank | Razorpay → Indian bank |
| Setup | Profile, bank, tax info, approval | KYC, bank, PAN, payment page |
| First payout | Often ~60 days after the first sponsorship | Normal Razorpay settlement cycle |

Neither one requires exposing a bank account or UPI ID on the website.

## 1. GitHub Sponsors

### Apply

1. Open <https://github.com/sponsors/accounts>.
2. Choose **Get sponsored** for the `MasterShivnath007` personal account.
3. Complete the profile:
   - Display name and short bio
   - Website: `https://mastershivnath007.github.io/dronagiri/`
   - Country: India
   - Bank account for payouts (via Stripe)
   - PAN/tax information
4. Accept the GitHub Sponsors terms and submit.

India is supported. Approval can take a few days. The first payout is usually
remitted about 60 days after the first sponsorship, then monthly — plan for that
delay.

### Tiers

Start low. A new project converts better with a $3 tier than a $25 one.

| Tier | Price | What it means |
|---|---|---|
| Supporter | **$3 / month** | Maintenance and release builds |
| Backer | **$10 / month** | Priority issue triage and roadmap vote |
| Company | **$50 / month** | Credit in the README and priority fixes |

Also enable **one-time sponsorship** amounts of $5, $25 and $100.

Set a concrete goal, for example:

> Cover the $99/year Apple Developer membership so releases can be signed and
> notarized.

### After approval

The profile becomes:

```
https://github.com/sponsors/MasterShivnath007
```

`.github/FUNDING.yml` already points there, so GitHub will show the Sponsor
button on the repository automatically. Send me the URL once it is live.

### Why the Sponsor button can still be disabled

`FUNDING.yml` is only a pointer. It does not create a donation page, a tier or a
button. The button stays disabled until **all** of these are true:

1. The GitHub Sponsors profile is approved and published.
2. At least one **monthly tier** is published with a price.
3. The profile is public.

Create the tiers at <https://github.com/sponsors/accounts> → your account →
**Tiers** (or **Sponsor tiers**) → **Add tier**:

- Supporter — $3/month
- Backer — $10/month
- Company — $50/month

Then enable **one-time amounts** ($5, $25, $100) and add the goal. After saving,
open the public page in a private window:

```
https://github.com/sponsors/MasterShivnath007
```

If it shows “Select a tier” with no prices, the tiers are not published yet. The
repository button will stay disabled until they are. Do not edit `FUNDING.yml`
to fix this; the file is already correct.

## 2. Razorpay

Use Razorpay for Indian supporters and for the paid audit. It handles UPI,
cards, netbanking and wallets, and pays into your bank account.

### Create the account

1. Sign up at <https://razorpay.com/> with your email and phone.
2. Choose the account type that fits you (individual/sole proprietor is fine for
   a one-person project).
3. Complete KYC:
   - PAN
   - Bank account details
   - Business/website details
   - Address proof if requested
4. Use the project site as the website:
   `https://mastershivnath007.github.io/dronagiri/`
5. Describe the business as software development / developer tools.

Razorpay may take a day or two to activate. Do not publish payment links until
KYC is approved.

### Create two Payment Pages

In the Razorpay Dashboard, open **Payment Pages** (or **Payment Links**) and
create:

#### A. Support Dronagiri — customer decides

- Title: `Support Dronagiri`
- Description: `Support the maintenance of the Dronagiri macOS package manager.`
- Amount: **customer decides**, minimum ₹100
- Collect: name and email
- Enable: UPI, cards, netbanking, wallets
- Success message: `Thank you. You are keeping Dronagiri maintained.`

Copy the public page URL.

#### B. Mac environment audit — fixed price

- Title: `Mac environment audit`
- Description: `A 60-minute remote audit using Dronagiri, with a prioritised cleanup report and seven days of follow-up.`
- Amount: fixed at the rupee equivalent of **$49** (for example ₹4,199)
- Collect: name, email, and a short note about the Mac
- Enable the same payment methods

Copy the public page URL.

If you later want international cards through Razorpay, enable international
payments in the account settings; otherwise GitHub Sponsors covers international
supporters.

### Test it

- Use Razorpay **Test Mode** first and make a test payment.
- Then make one small live payment (₹1–₹10) and refund it, or ask a friend to pay
  ₹1, to confirm the page and settlement work.
- Check the payout/settlement schedule in the dashboard.

### Fees and tax

Razorpay charges a processing fee (roughly 2% for domestic cards; UPI and
international cards differ). Check the current pricing page before setting
prices.

Sponsorship and donation income is generally taxable in India. Keep a simple
record of payouts and talk to a CA when the amounts become meaningful. This is
not legal or tax advice.

## The website support page

`https://mastershivnath007.github.io/dronagiri/#support` is not a GitHub
Sponsors page and not a payment page. It is the **Support** section of the
project website. It contains:

- a link to GitHub Sponsors
- an optional paid **Mac environment audit — $49**
- a link to the full support policy

The audit is a service an individual can book, not a fee for using Dronagiri.
The tool stays free and MIT-licensed. If you do not want to offer the audit,
say so and I will remove the $49 card from `docs/index.html`, `README.md` and
`SUPPORT.md`. If you want a different price or scope, tell me and I will change
it everywhere.

## Live links

| Purpose | URL |
|---|---|
| GitHub Sponsors | <https://github.com/sponsors/MasterShivnath007> |
| One-time support (Razorpay) | <https://rzp.io/rzp/gkf8SP9> |
| Mac audit (Razorpay) | <https://rzp.io/rzp/PWulBM7> |

GitHub Sponsors is active with six tiers ($3, $10 and $50 monthly; $5, $25 and
$100 one-time), and the repository Sponsor button is enabled.

The Razorpay **Support Dronagiri** page accepts a customer-entered amount with a
₹100 minimum. That is correct for a donation page.

The Razorpay **Mac environment audit** page is also currently customer-entered.
Before sharing it as a $49 service, edit the page in Razorpay and set a **fixed
amount of ₹4,199** (the rupee equivalent of $49). Until then, the website keeps
the audit request on the GitHub issue form so nobody can underpay by mistake.
Once the fixed amount is saved, send the updated link and the site button will
point at it.

The payment pages currently show a personal email and phone number. If you do
not want those public, change the contact details in Razorpay to a project email
before sharing the pages widely.

## What is already wired

- Website support section: GitHub Sponsors + Razorpay one-time + support policy
- README support section: GitHub Sponsors + Razorpay + audit request
- `.github/FUNDING.yml`: GitHub Sponsors + Razorpay one-time
- `SUPPORT.md`: live tier prices and both payment links

## Do not

- Do not post your bank account number, PAN, or personal UPI ID publicly.
- Do not use a Razorpay page before KYC is complete; payments can be held.
- Do not promise rewards you cannot deliver. A refund is cheaper than an angry
  supporter.
- Do not call it a “charity” or “donation” in legal/tax contexts; use
  “sponsorship” or “support” for an individual project.

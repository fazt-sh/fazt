Supabase has seen explosive growth in 2024 and 2025, largely driven by the "AI-generated app" (Vibe Coding) boom. For someone building an alternative, the key is understanding that Supabase is no longer just a "database"; it is an **operating system for the long tail of developers.**

Below is a deep research breakdown of their current stats and realistic estimates for their internal metrics as of late 2025.

### 1. High-Level Adoption Stats (Estimates)

Supabase does not release exact internal dashboard numbers, but based on revenue reports (Sacra, TapTwice) and infrastructure footprints, we can triangulate the following:

| Metric | Estimate (Late 2025) | Notes |
| --- | --- | --- |
| **Total Projects (DBs)** | **1.2 Million - 1.5 Million** | Sacra reports they hit 1M databases mid-2024. |
| **Active Projects** | **150,000 - 250,000** | ~15-20% of total projects see weekly traffic. |
| **Total Registered Users** | **800,000 - 1.2 Million** | Many users have multiple projects; high churn in hobbyists. |
| **Paying Customers** | **40,000 - 60,000** | Based on $70M ARR and typical SaaS conversion (4-6%). |
| **GitHub Stars** | **~75,000** | A core proxy for developer "mindshare." |

---

### 2. Individual User Spending (Average vs. Median)

The "Individual User" category on Supabase is bifurcated between **Hobbyists** (Free) and **Indie Hackers/Prosumers** (Pro).

* **The Median Spend: $0**
The vast majority of users never leave the Free tier. They use Supabase for prototypes, side projects, or "sleeping" apps that get paused after a week of inactivity.
* **The Pro Median Spend: $25 - $35/month**
For a paying user, the floor is $25. Most Pro users run 1-2 projects. Since the first $10 of compute is credited back, a single-project Pro user pays exactly $25. If they have a second small project, it bumps to $35 ($25 base + $10 compute).
* **The Average Spend (ARPU of Paying Users): ~$100 - $120/month**
While most pay $25, a small "whale" population of startups pays $600+ (Team Plan) or thousands in overages (egress/storage). This pulls the average significantly higher than the median.

---

### 3. Usage Stats of a "Typical" App

To compete, your alternative must handle the **"Long Tail" distribution**: most apps are tiny, but 1% will consume 90% of your resources.

#### **Project Profiles**

| Metric | Median (The "Average" App) | 95th Percentile (The "Scalers") |
| --- | --- | --- |
| **Database Size** | **< 100 MB** | **10 GB+** |
| **Monthly Egress** | **< 500 MB** | **50 GB+** |
| **Storage (Files)** | **< 50 MB** | **5 GB+** |
| **Monthly Active Users** | **< 100 MAU** | **10,000+ MAU** |
| **Realtime Conns** | **1-5 concurrent** | **500+ concurrent** |

> **Analyst Insight:** Most Supabase users are building "CRUD" apps where the database is mostly empty. The primary value they get isn't "big data" storage; it's the **convenience** of Auth + API generation.

---

### 4. Holistic View: Why are people using it?

If you are building an alternative, you aren't just competing on database performance. You are competing on **ecosystem integration.**

1. **The "Vibe Coding" Default:** In 2025, Supabase became the default backend for AI coding tools like **Bolt.new** and **Lovable**. These tools auto-generate Supabase schemas. If your alternative doesn't have an "adapter" for these AI agents, you miss the biggest growth engine of the year.
2. **The "Postgres Maxis":** Users choose Supabase because it *is* Postgres. They want the ability to leave. If your alternative uses a proprietary or "Postgres-compatible" (but not true Postgres) engine, you lose the "no vendor lock-in" crowd.
3. **The Auth/Storage/Edge Hook:** Most users start for the Database but stay for the **Auth (RLS)**. Row Level Security is the "killer feature" that makes frontend-only development possible.

### 5. Estimates for Your Business Model

If you want your alternative to "hold up," plan for these infrastructure realities:

* **Inactivity Management:** You must be able to "hibernate" (pause) projects. If you keep 1 million Postgres instances "hot," your cloud bill will kill your startup before you reach $1M revenue.
* **Bandwidth is the Margin Killer:** Supabase charges $0.09/GB for egress overages. This is where they make their highest margin. To compete, you could offer "Free Egress up to 100GB" to lure their Pro users away.

**Would you like me to draft a competitive "Feature Gap Analysis" between Supabase and your current architectural plan?**
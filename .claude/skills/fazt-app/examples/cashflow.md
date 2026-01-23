# CashFlow - Reference Example

**Type**: Expense Tracker
**Complexity**: Advanced
**Storage**: Document Store (ds), Key-Value (kv)
**Live**: https://cashflow.zyt.app

## What It Demonstrates

### Advanced Storage Patterns
- **Complex queries**: Session-scoped, compound queries, date ranges
- **$inc operator**: Running totals for category budgets
- **Cascading updates**: When transaction changes, category totals update
- **Relationships**: Transactions linked to categories
- **Aggregations**: Multi-query stats with JavaScript grouping
- **KV caching**: Stats cached with 5-minute TTL

### UX Excellence
- **Fixed-height layout**: Prevents scroll jumps with `h-screen flex flex-col`
- **Scrollbar-gutter**: Stable scrollbar space prevents layout shift
- **Click-outside modals**: Both `@click.self` and backdrop `@click` handlers
- **Keyboard shortcuts**: Escape to close modals
- **Sound effects**: Success, tap, error tones
- **Theme system**: Light/Dark/System with proper toggle
- **Session management**: URL-based 3-word sessions

### Features
- 3 views: Transactions, Categories, Stats
- Date filtering: Today, Week, Month, Year, All Time
- Category budgets with visual progress bars
- Income & Expense tracking
- Auto-initialization with 12 default categories
- Currency selector

## Key Files

### Layout (index.html)
```css
/* Fixed height prevents scroll jumps */
html, body { height: 100%; overflow: hidden; }
#app { height: 100%; }

/* Prevent scrollbar layout shift */
.overflow-y-auto { scrollbar-gutter: stable; }
```

### Main App (src/main.js)
- Flexbox layout: `h-screen flex flex-col overflow-hidden`
- Header: `flex-none` (fixed)
- Content: `flex-1 overflow-y-auto` (scrollable)
- FAB button: Fixed outside scroll container

### API (api/main.js)
- Session-scoped CRUD operations
- Transaction → Category relationship
- Real-time budget calculations with `$inc`
- Stats endpoint with multiple aggregations

## Storage Operations

### Insert with Relationship
```javascript
// Create transaction
ds.insert('transactions', { id, session, type, amount, category, date })

// Update category total
ds.update('categories', { session, name: category }, {
  $inc: { totalSpent: amount }
})
```

### Complex Query
```javascript
// Date range with comparison
var monthStartStr = monthStart.toISOString().split('T')[0]
var txns = ds.find('transactions', {
  session: session,
  date: { $gte: monthStartStr }
})
```

### Cascading Update
```javascript
// When transaction changes, adjust both old and new category totals
ds.update('categories', { session, name: oldCategory }, {
  $inc: { totalSpent: -oldAmount }
})
ds.update('categories', { session, name: newCategory }, {
  $inc: { totalSpent: newAmount }
})
```

## Lessons Learned

1. **Always use fixed-height layouts** for apps with dynamic content
2. **Double-check modal close handlers** - both backdrop and self
3. **Initialize default data** when collections are empty
4. **Use $inc for counters** instead of read-modify-write
5. **Test with FAZT_DEBUG=1** to see storage operations in real-time
6. **Smart type switching** - auto-select matching categories when type changes

## Performance

With debug mode, typical operation timings:
```
[DEBUG storage] insert cashflow/transactions rows=1 took=1.2ms
[DEBUG storage] find cashflow/categories query={session:"..."} rows=12 took=0.8ms
[DEBUG storage] update cashflow/categories rows=1 took=1.3ms
[DEBUG runtime] req=a1b2c3 app=cashflow path=/api/transactions status=201 took=4.5ms
```

## Source

Check `servers/zyt/cashflow/` for full source (gitignored, local only).

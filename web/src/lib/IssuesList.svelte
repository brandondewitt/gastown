<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  interface BeadItem {
    id: string;
    title: string;
    type: string;
    status: string;
    priority: number;
    priority_str?: string;
    created_at?: string;
    updated_at?: string;
  }

  interface Props {
    onDispatch?: (issueId: string) => void;
  }

  let { onDispatch }: Props = $props();

  // State
  let issues: BeadItem[] = $state([]);
  let isLoading: boolean = $state(true);
  let error: string | null = $state(null);
  let statusFilter: string = $state('open');
  let typeFilter: string = $state('');
  let expandedId: string | null = $state(null);
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  // Fetch issues on mount and set up auto-refresh
  onMount(() => {
    fetchIssues();
    refreshInterval = setInterval(fetchIssues, 30000);
  });

  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });

  async function fetchIssues() {
    try {
      isLoading = true;
      error = null;

      const params = new URLSearchParams();
      if (statusFilter) {
        params.set('status', statusFilter);
      }
      if (typeFilter) {
        params.set('type', typeFilter);
      }

      const url = `/api/v1/beads${params.toString() ? '?' + params.toString() : ''}`;
      const response = await fetch(url);
      const data = await response.json();

      if (data.success) {
        issues = data.data || [];
      } else {
        error = data.error?.message || 'Failed to load issues';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load issues';
    } finally {
      isLoading = false;
    }
  }

  function handleStatusChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    statusFilter = target.value;
    fetchIssues();
  }

  function handleTypeChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    typeFilter = target.value;
    fetchIssues();
  }

  function toggleExpanded(id: string) {
    expandedId = expandedId === id ? null : id;
  }

  function handleDispatch(issueId: string) {
    if (onDispatch) {
      onDispatch(issueId);
    }
  }

  function getTypeColor(type: string): string {
    switch (type) {
      case 'bug': return 'var(--color-error)';
      case 'feature': return 'var(--color-primary)';
      case 'epic': return '#8b5cf6';
      case 'chore': return 'var(--color-text-muted)';
      default: return 'var(--color-success)';
    }
  }

  function getPriorityLabel(priority: number): string {
    switch (priority) {
      case 0: return 'P0';
      case 1: return 'P1';
      case 2: return 'P2';
      case 3: return 'P3';
      case 4: return 'P4';
      default: return `P${priority}`;
    }
  }

  function getPriorityColor(priority: number): string {
    switch (priority) {
      case 0: return '#ef4444';
      case 1: return '#f97316';
      case 2: return '#eab308';
      case 3: return '#22c55e';
      case 4: return 'var(--color-text-muted)';
      default: return 'var(--color-text-muted)';
    }
  }

  function formatDate(dateStr?: string): string {
    if (!dateStr) return '';
    try {
      const date = new Date(dateStr);
      return date.toLocaleDateString();
    } catch {
      return dateStr;
    }
  }
</script>

<div class="issues-list">
  <div class="header">
    <h2>Issues</h2>
    <button class="refresh-btn" onclick={fetchIssues} disabled={isLoading}>
      {isLoading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  <div class="filters">
    <div class="filter-group">
      <label for="status-filter">Status</label>
      <select id="status-filter" value={statusFilter} onchange={handleStatusChange}>
        <option value="open">Open</option>
        <option value="closed">Closed</option>
        <option value="all">All</option>
      </select>
    </div>

    <div class="filter-group">
      <label for="type-filter">Type</label>
      <select id="type-filter" value={typeFilter} onchange={handleTypeChange}>
        <option value="">All Types</option>
        <option value="task">Task</option>
        <option value="bug">Bug</option>
        <option value="feature">Feature</option>
        <option value="epic">Epic</option>
        <option value="chore">Chore</option>
      </select>
    </div>
  </div>

  {#if error}
    <div class="error-message">
      {error}
    </div>
  {/if}

  {#if isLoading && issues.length === 0}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading issues...</p>
    </div>
  {:else if issues.length === 0}
    <div class="empty-state">
      <p>No issues found</p>
    </div>
  {:else}
    <div class="issues">
      {#each issues as issue (issue.id)}
        <div class="issue-card" class:expanded={expandedId === issue.id}>
          <div class="issue-header" onclick={() => toggleExpanded(issue.id)}>
            <div class="issue-id">{issue.id}</div>
            <div class="issue-title">{issue.title}</div>
            <div class="issue-badges">
              <span class="badge type-badge" style="background-color: {getTypeColor(issue.type)}">
                {issue.type}
              </span>
              <span class="badge priority-badge" style="background-color: {getPriorityColor(issue.priority)}">
                {getPriorityLabel(issue.priority)}
              </span>
              <span class="badge status-badge" class:open={issue.status === 'open'} class:closed={issue.status === 'closed'}>
                {issue.status}
              </span>
            </div>
            <div class="expand-icon">{expandedId === issue.id ? '▼' : '▶'}</div>
          </div>

          {#if expandedId === issue.id}
            <div class="issue-details">
              <div class="detail-row">
                <span class="detail-label">Created:</span>
                <span class="detail-value">{formatDate(issue.created_at) || 'N/A'}</span>
              </div>
              {#if issue.updated_at}
                <div class="detail-row">
                  <span class="detail-label">Updated:</span>
                  <span class="detail-value">{formatDate(issue.updated_at)}</span>
                </div>
              {/if}
              <div class="issue-actions">
                <button class="action-btn dispatch-btn" onclick={() => handleDispatch(issue.id)}>
                  Dispatch
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <div class="footer">
    <span class="count">{issues.length} issue{issues.length !== 1 ? 's' : ''}</span>
  </div>
</div>

<style>
  .issues-list {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header h2 {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
  }

  .refresh-btn {
    padding: var(--space-1) var(--space-3);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    color: var(--color-text);
    font-size: 0.875rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .refresh-btn:hover:not(:disabled) {
    background-color: var(--color-border);
  }

  .refresh-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .filters {
    display: flex;
    gap: var(--space-4);
    flex-wrap: wrap;
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .filter-group label {
    font-size: 0.875rem;
    color: var(--color-text-muted);
  }

  .filter-group select {
    padding: var(--space-1) var(--space-2);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    color: var(--color-text);
    font-size: 0.875rem;
  }

  .error-message {
    padding: var(--space-3);
    background-color: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: var(--radius-sm);
    color: var(--color-error);
    font-size: 0.875rem;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: var(--space-8);
    gap: var(--space-3);
  }

  .spinner {
    width: 2rem;
    height: 2rem;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading p {
    color: var(--color-text-muted);
    font-size: 0.875rem;
  }

  .empty-state {
    text-align: center;
    padding: var(--space-8);
    color: var(--color-text-muted);
  }

  .issues {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .issue-card {
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    overflow: hidden;
    transition: border-color 0.2s;
  }

  .issue-card:hover {
    border-color: var(--color-primary);
  }

  .issue-header {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    cursor: pointer;
  }

  .issue-id {
    font-family: var(--font-mono, monospace);
    font-size: 0.75rem;
    color: var(--color-text-muted);
    min-width: 80px;
  }

  .issue-title {
    flex: 1;
    font-size: 0.875rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .issue-badges {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
  }

  .badge {
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.7rem;
    font-weight: 500;
    text-transform: uppercase;
    color: white;
  }

  .status-badge {
    background-color: var(--color-text-muted);
  }

  .status-badge.open {
    background-color: var(--color-success);
  }

  .status-badge.closed {
    background-color: var(--color-text-muted);
  }

  .expand-icon {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    width: 16px;
    text-align: center;
  }

  .issue-details {
    padding: var(--space-3);
    padding-top: 0;
    border-top: 1px solid var(--color-border);
  }

  .detail-row {
    display: flex;
    gap: var(--space-2);
    font-size: 0.875rem;
    padding: var(--space-1) 0;
  }

  .detail-label {
    color: var(--color-text-muted);
    min-width: 80px;
  }

  .detail-value {
    color: var(--color-text);
  }

  .issue-actions {
    display: flex;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }

  .action-btn {
    padding: var(--space-1) var(--space-3);
    border: none;
    border-radius: var(--radius-sm);
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .dispatch-btn {
    background-color: var(--color-primary);
    color: white;
  }

  .dispatch-btn:hover {
    background-color: #2563eb;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
    padding-top: var(--space-2);
    border-top: 1px solid var(--color-border);
  }

  .count {
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }
</style>

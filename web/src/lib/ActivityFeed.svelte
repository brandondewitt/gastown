<script lang="ts">
  import { onMount } from 'svelte';

  interface FeedEvent {
    id: string;
    type: string;
    timestamp: string;
    actor?: string;
    rig?: string;
    message: string;
    details?: unknown;
  }

  interface PaginatedResponse {
    items: FeedEvent[];
    total: number;
    offset: number;
    limit: number;
    has_more: boolean;
  }

  // State
  let events = $state<FeedEvent[]>([]);
  let loading = $state(true);
  let error: string | null = $state(null);
  let feedContainer: HTMLDivElement | undefined;

  // Refresh interval
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  onMount(() => {
    // Initial fetch
    fetchEvents();

    // Set up auto-refresh every 10 seconds
    refreshInterval = setInterval(() => {
      fetchEvents();
    }, 10000);

    // Cleanup function
    return () => {
      if (refreshInterval) clearInterval(refreshInterval);
    };
  });

  async function fetchEvents() {
    try {
      error = null;
      const response = await fetch('/api/v1/events?limit=50');
      const data = await response.json();

      if (data.success && data.data) {
        const paginated = data.data as PaginatedResponse;
        events = paginated.items || [];
      } else {
        error = data.error?.message || 'Failed to load events';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load events';
    } finally {
      loading = false;
    }
  }

  function getEventIcon(type: string): string {
    const icons: Record<string, string> = {
      created: '✨',
      updated: '✏️',
      deleted: '🗑️',
      started: '▶️',
      completed: '✓',
      failed: '❌',
      assigned: '🎯',
      merged: '🔀',
      dispatched: '🚀',
      hook: '🪝',
      mail: '📧',
      error: '⚠️',
      warning: '⚠️',
      info: 'ℹ️',
      default: '📌',
    };
    return icons[type.toLowerCase()] || icons.default;
  }

  function getEventColor(type: string): string {
    const colors: Record<string, string> = {
      created: 'var(--color-success)',
      updated: 'var(--color-primary)',
      deleted: 'var(--color-error)',
      started: 'var(--color-primary)',
      completed: 'var(--color-success)',
      failed: 'var(--color-error)',
      assigned: 'var(--color-primary)',
      merged: 'var(--color-success)',
      dispatched: 'var(--color-primary)',
      hook: 'var(--color-primary)',
      mail: 'var(--color-primary)',
      error: 'var(--color-error)',
      warning: 'var(--color-warning)',
      info: 'var(--color-primary)',
    };
    return colors[type.toLowerCase()] || 'var(--color-gray)';
  }

  function formatTimestamp(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffSecs = Math.floor(diffMs / 1000);
      const diffMins = Math.floor(diffSecs / 60);
      const diffHours = Math.floor(diffMins / 60);
      const diffDays = Math.floor(diffHours / 24);

      if (diffSecs < 60) {
        return 'just now';
      } else if (diffMins < 60) {
        return `${diffMins}m ago`;
      } else if (diffHours < 24) {
        return `${diffHours}h ago`;
      } else if (diffDays < 7) {
        return `${diffDays}d ago`;
      } else {
        return date.toLocaleDateString(undefined, {
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
        });
      }
    } catch {
      return timestamp;
    }
  }
</script>

<div class="activity-feed">
  <div class="feed-header">
    <h2 class="feed-title">Activity Feed</h2>
    <button
      class="refresh-btn"
      onclick={fetchEvents}
      disabled={loading}
      title="Refresh events"
      aria-label="Refresh activity feed"
    >
      {#if loading}
        <span class="spinner-small"></span>
      {:else}
        ↻
      {/if}
    </button>
  </div>

  <div class="feed-container" bind:this={feedContainer}>
    {#if loading && events.length === 0}
      <div class="feed-state loading">
        <div class="spinner"></div>
        <p>Loading activity...</p>
      </div>
    {:else if error}
      <div class="feed-state error">
        <p class="error-message">⚠️ {error}</p>
        <button class="btn-retry" onclick={fetchEvents}>
          Try Again
        </button>
      </div>
    {:else if events.length === 0}
      <div class="feed-state empty">
        <p>No events yet</p>
        <p class="text-muted">Events will appear as they happen</p>
      </div>
    {:else}
      <div class="events-list">
        {#each events as event (event.id)}
          <div class="event-item" role="article">
            <div class="event-icon" style="color: {getEventColor(event.type)}">
              {getEventIcon(event.type)}
            </div>
            <div class="event-content">
              <div class="event-header">
                <span class="event-type">{event.type}</span>
                {#if event.rig}
                  <span class="event-rig">{event.rig}</span>
                {/if}
                <span class="event-time">{formatTimestamp(event.timestamp)}</span>
              </div>
              <p class="event-message">{event.message}</p>
              {#if event.actor}
                <p class="event-actor mono text-muted">by {event.actor}</p>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .activity-feed {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .feed-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4);
    border-bottom: 1px solid var(--color-border);
    flex-shrink: 0;
  }

  .feed-title {
    font-size: 1rem;
    font-weight: 500;
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0;
  }

  .refresh-btn {
    background: none;
    border: none;
    color: var(--color-text-muted);
    font-size: 1rem;
    cursor: pointer;
    padding: var(--space-2);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  .refresh-btn:hover:not(:disabled) {
    background-color: var(--color-surface-raised);
    color: var(--color-text);
  }

  .refresh-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .feed-container {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  .feed-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-8);
    text-align: center;
    flex: 1;
  }

  .feed-state.loading {
    gap: var(--space-4);
  }

  .feed-state.empty {
    gap: var(--space-2);
    color: var(--color-text-muted);
  }

  .feed-state.error {
    gap: var(--space-4);
  }

  .error-message {
    color: var(--color-error);
    margin: 0;
  }

  .btn-retry {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-sm);
    font-weight: 500;
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }

  .btn-retry:hover {
    background-color: #2563eb;
  }

  .events-list {
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .event-item {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-4);
    border-bottom: 1px solid var(--color-border);
    transition: background-color var(--transition-fast);
  }

  .event-item:last-child {
    border-bottom: none;
  }

  .event-item:hover {
    background-color: var(--color-surface-raised);
  }

  .event-icon {
    font-size: 1.25rem;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
  }

  .event-content {
    flex: 1;
    min-width: 0;
  }

  .event-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-1);
    flex-wrap: wrap;
  }

  .event-type {
    font-weight: 600;
    color: var(--color-text);
    text-transform: capitalize;
  }

  .event-rig {
    font-size: 0.75rem;
    padding: 0.125rem var(--space-1);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-sm);
    color: var(--color-text-muted);
    font-family: var(--font-mono);
  }

  .event-time {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    margin-left: auto;
  }

  .event-message {
    color: var(--color-text);
    margin: 0 0 var(--space-1) 0;
    word-break: break-word;
  }

  .event-actor {
    font-size: 0.75rem;
    margin: 0;
  }

  .spinner {
    width: 1.5rem;
    height: 1.5rem;
    border: 2px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .spinner-small {
    display: inline-block;
    width: 0.875rem;
    height: 0.875rem;
    border: 2px solid rgba(59, 130, 246, 0.3);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* Scrollbar styling */
  .feed-container::-webkit-scrollbar {
    width: 6px;
  }

  .feed-container::-webkit-scrollbar-track {
    background: transparent;
  }

  .feed-container::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: 3px;
  }

  .feed-container::-webkit-scrollbar-thumb:hover {
    background: var(--color-gray);
  }
</style>

<script lang="ts">
  import { onMount } from 'svelte';

  interface MergeQueueItem {
    id: string;
    branch: string;
    status: 'pending' | 'in_flight' | 'blocked' | 'completed' | 'failed';
    position: number;
    created_at: string;
  }

  interface ApiResponse {
    success: boolean;
    data?: MergeQueueItem[];
    error?: {
      code: string;
      message: string;
    };
  }

  interface RetryResponse {
    success: boolean;
    message?: string;
    error?: {
      code: string;
      message: string;
    };
  }

  let items: MergeQueueItem[] = $state([]);
  let loading: boolean = $state(true);
  let error: string | null = $state(null);
  let retryingId: string | null = $state(null);

  async function fetchMergeQueue() {
    try {
      loading = true;
      error = null;
      const response = await fetch('/api/v1/mq');
      const data: ApiResponse = await response.json();

      if (data.success) {
        items = data.data || [];  // Handle null as empty array
      } else {
        error = data.error?.message || 'Failed to fetch merge queue';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to fetch merge queue';
    } finally {
      loading = false;
    }
  }

  async function retryItem(id: string) {
    try {
      retryingId = id;
      const response = await fetch(`/api/v1/mq/${id}/retry`, {
        method: 'POST',
      });
      const data: RetryResponse = await response.json();

      if (data.success) {
        // Refresh the queue after successful retry
        await fetchMergeQueue();
      } else {
        error = data.error?.message || 'Failed to retry merge';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to retry merge';
    } finally {
      retryingId = null;
    }
  }

  function getStatusColor(status: string): string {
    const colors: Record<string, string> = {
      pending: 'var(--color-warning)',
      in_flight: 'var(--color-info)',
      blocked: 'var(--color-error)',
      completed: 'var(--color-success)',
      failed: 'var(--color-error)',
    };
    return colors[status] || 'var(--color-gray)';
  }

  function getStatusIcon(status: string): string {
    const icons: Record<string, string> = {
      pending: '⏳',
      in_flight: '🔄',
      blocked: '🚧',
      completed: '✅',
      failed: '❌',
    };
    return icons[status] || '❓';
  }

  function formatTime(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return 'just now';
  }

  onMount(() => {
    fetchMergeQueue();
    // Refresh every 10 seconds
    const interval = setInterval(fetchMergeQueue, 10000);
    return () => clearInterval(interval);
  });
</script>

<div class="merge-queue">
  {#if loading && items.length === 0}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading merge queue...</p>
    </div>
  {:else if error && items.length === 0}
    <div class="error-container">
      <p class="text-error">Error: {error}</p>
      <button on:click={fetchMergeQueue}>Retry</button>
    </div>
  {:else if items.length === 0}
    <div class="empty-state">
      <p>Merge queue is empty</p>
    </div>
  {:else}
    <div class="queue-container">
      <h3 class="queue-title">Merge Queue ({items.length} items)</h3>
      {#if error}
        <div class="error-banner">
          <p>{error}</p>
        </div>
      {/if}
      <div class="queue-list">
        {#each items as item (item.id)}
          <div class="queue-item" class:failed={item.status === 'failed'}>
            <div class="item-position">
              <span class="position-number">#{item.position}</span>
            </div>
            <div class="item-content">
              <div class="branch-info">
                <span class="branch-name">{item.branch}</span>
                <span class="status-badge" style="background-color: {getStatusColor(item.status)}">
                  <span class="status-icon">{getStatusIcon(item.status)}</span>
                  <span class="status-text">{item.status}</span>
                </span>
              </div>
              <div class="item-meta">
                <span class="time-in-queue">Enqueued {formatTime(item.created_at)}</span>
              </div>
            </div>
            {#if item.status === 'failed'}
              <div class="item-actions">
                <button
                  class="retry-button"
                  on:click={() => retryItem(item.id)}
                  disabled={retryingId === item.id}
                >
                  {retryingId === item.id ? 'Retrying...' : 'Retry'}
                </button>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .merge-queue {
    width: 100%;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-8);
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-8);
  }

  .error-container button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
  }

  .empty-state {
    text-align: center;
    padding: var(--space-8);
    color: var(--color-text-muted);
  }

  .queue-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .queue-title {
    font-size: 1.125rem;
    font-weight: 600;
    margin: 0;
  }

  .error-banner {
    padding: var(--space-3) var(--space-4);
    background-color: var(--color-error-surface);
    border-left: 3px solid var(--color-error);
    border-radius: var(--radius-md);
    color: var(--color-error);
  }

  .error-banner p {
    margin: 0;
    font-size: 0.875rem;
  }

  .queue-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .queue-item {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    transition: all 0.2s ease;
  }

  .queue-item:hover {
    border-color: var(--color-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .queue-item.failed {
    border-color: var(--color-error);
    background-color: var(--color-error-surface);
  }

  .item-position {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 50px;
  }

  .position-number {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-primary);
  }

  .item-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .branch-info {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }

  .branch-name {
    font-weight: 500;
    color: var(--color-text);
    word-break: break-all;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    color: white;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-icon {
    font-size: 0.875rem;
  }

  .status-text {
    display: inline;
  }

  .item-meta {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    font-size: 0.875rem;
    color: var(--color-text-muted);
  }

  .time-in-queue {
    display: flex;
    align-items: center;
  }

  .item-actions {
    display: flex;
    gap: var(--space-2);
  }

  .retry-button {
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-warning);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .retry-button:hover:not(:disabled) {
    background-color: var(--color-warning-dark);
  }

  .retry-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  @media (max-width: 640px) {
    .queue-item {
      flex-direction: column;
      align-items: flex-start;
    }

    .item-position {
      min-width: auto;
      margin-bottom: var(--space-2);
    }

    .branch-info {
      width: 100%;
    }

    .item-actions {
      width: 100%;
      margin-top: var(--space-2);
    }

    .retry-button {
      width: 100%;
    }
  }
</style>

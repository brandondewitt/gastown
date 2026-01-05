<script lang="ts">
  import { onMount } from 'svelte';

  // Types matching the backend Message struct
  interface Message {
    id: string;
    from: string;
    to: string;
    subject: string;
    body: string;
    timestamp: string;
    read: boolean;
    priority: 'low' | 'normal' | 'high' | 'urgent';
    type: 'task' | 'scavenge' | 'notification' | 'reply';
    thread_id?: string;
    reply_to?: string;
  }

  export let agentAddress: string = '';

  let messages: Message[] = [];
  let loading = false;
  let error: string | null = null;
  let expandedMessages: Set<string> = new Set();
  let showComposer = false;
  let replyingTo: Message | null = null;

  async function loadMessages() {
    if (!agentAddress) {
      error = 'Agent address required';
      return;
    }

    loading = true;
    error = null;

    try {
      const response = await fetch(`/api/v1/mail?agent=${encodeURIComponent(agentAddress)}`);
      const data = await response.json();

      if (data.success) {
        messages = data.data || [];
      } else {
        error = data.error?.message || 'Failed to load messages';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load messages';
    } finally {
      loading = false;
    }
  }

  async function deleteMessage(id: string) {
    if (!confirm('Delete this message?')) {
      return;
    }

    try {
      const response = await fetch(
        `/api/v1/mail/${id}?agent=${encodeURIComponent(agentAddress)}`,
        { method: 'DELETE' }
      );
      const data = await response.json();

      if (data.success) {
        messages = messages.filter((m) => m.id !== id);
      } else {
        alert(data.error?.message || 'Failed to delete message');
      }
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Failed to delete message');
    }
  }

  function toggleExpanded(id: string) {
    if (expandedMessages.has(id)) {
      expandedMessages.delete(id);
    } else {
      expandedMessages.add(id);
    }
    expandedMessages = expandedMessages; // Trigger reactivity
  }

  function openReplyComposer(msg: Message) {
    replyingTo = msg;
    showComposer = true;
  }

  function closeComposer() {
    showComposer = false;
    replyingTo = null;
  }

  function getPriorityColor(priority: string): string {
    const colors: Record<string, string> = {
      urgent: '#ef4444',
      high: '#f97316',
      normal: '#6b7280',
      low: '#9ca3af',
    };
    return colors[priority] || '#6b7280';
  }

  function formatTimestamp(ts: string): string {
    const date = new Date(ts);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  onMount(() => {
    if (agentAddress) {
      loadMessages();
    }
  });
</script>

<div class="mail-container">
  <div class="mail-header">
    <h2>Mail</h2>
    {#if messages.length > 0}
      <span class="message-count">{messages.length}</span>
    {/if}
    <button class="refresh-btn" on:click={loadMessages} disabled={loading}>
      {loading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  {#if loading && messages.length === 0}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading mail...</p>
    </div>
  {:else if error && messages.length === 0}
    <div class="error">
      <p>{error}</p>
      <button on:click={loadMessages}>Retry</button>
    </div>
  {:else if messages.length === 0}
    <div class="empty">
      <p>No messages</p>
    </div>
  {:else}
    <div class="message-list">
      {#each messages as msg (msg.id)}
        <div class="message-item" class:expanded={expandedMessages.has(msg.id)}>
          <div class="message-summary" on:click={() => toggleExpanded(msg.id)}>
            <div class="message-meta">
              <div class="priority-indicator" style="background-color: {getPriorityColor(msg.priority)}" title={msg.priority}></div>
              <div class="message-info">
                <div class="message-from">{msg.from}</div>
                <div class="message-subject">{msg.subject}</div>
              </div>
            </div>
            <div class="message-date">{formatTimestamp(msg.timestamp)}</div>
          </div>

          {#if expandedMessages.has(msg.id)}
            <div class="message-details">
              <div class="detail-row">
                <span class="label">From:</span>
                <span class="value">{msg.from}</span>
              </div>
              <div class="detail-row">
                <span class="label">To:</span>
                <span class="value">{msg.to}</span>
              </div>
              <div class="detail-row">
                <span class="label">Type:</span>
                <span class="value">{msg.type}</span>
              </div>
              <div class="message-body">{msg.body}</div>

              <div class="message-actions">
                <button class="btn-reply" on:click={() => openReplyComposer(msg)}>
                  ↩️ Reply
                </button>
                <button class="btn-delete" on:click={() => deleteMessage(msg.id)}>
                  🗑️ Delete
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  {#if showComposer && replyingTo}
    <div class="composer-overlay">
      <div class="composer-modal">
        <div class="composer-header">
          <h3>Reply to "{replyingTo.subject}"</h3>
          <button class="close-btn" on:click={closeComposer}>✕</button>
        </div>
        <div class="composer-body">
          <textarea placeholder="Type your reply..."></textarea>
        </div>
        <div class="composer-footer">
          <button class="btn-send">Send</button>
          <button class="btn-cancel" on:click={closeComposer}>Cancel</button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .mail-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-4);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-lg);
  }

  .mail-header {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    justify-content: space-between;
  }

  .mail-header h2 {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
  }

  .message-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 var(--space-1);
    font-size: 0.75rem;
    font-weight: 600;
    background-color: var(--color-primary);
    color: white;
    border-radius: 9999px;
  }

  .refresh-btn {
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.2s;
  }

  .refresh-btn:hover:not(:disabled) {
    background-color: var(--color-surface);
    border-color: var(--color-primary);
  }

  .refresh-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .loading,
  .error,
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-8);
    text-align: center;
  }

  .error {
    color: var(--color-error);
  }

  .error button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-error);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
  }

  .message-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .message-item {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .message-summary {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3);
    background-color: var(--color-surface);
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .message-summary:hover {
    background-color: var(--color-surface);
  }

  .message-item.expanded .message-summary {
    background-color: var(--color-surface);
    border-bottom: 1px solid var(--color-border);
  }

  .message-meta {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: 1;
  }

  .priority-indicator {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
  }

  .message-info {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    flex: 1;
    min-width: 0;
  }

  .message-from {
    font-weight: 500;
    color: var(--color-text);
  }

  .message-subject {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .message-date {
    font-size: 0.875rem;
    color: var(--color-text-muted);
    white-space: nowrap;
    margin-left: var(--space-4);
  }

  .message-details {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    background-color: var(--color-surface);
  }

  .detail-row {
    display: flex;
    gap: var(--space-2);
    font-size: 0.875rem;
  }

  .detail-row .label {
    font-weight: 500;
    color: var(--color-text-muted);
    min-width: 4rem;
  }

  .detail-row .value {
    color: var(--color-text);
    word-break: break-word;
  }

  .message-body {
    padding: var(--space-3);
    background-color: var(--color-surface-raised);
    border-radius: var(--radius-md);
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 0.875rem;
    line-height: 1.5;
    max-height: 300px;
    overflow-y: auto;
  }

  .message-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }

  .btn-reply,
  .btn-delete {
    padding: var(--space-2) var(--space-3);
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.2s;
  }

  .btn-reply {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-reply:hover {
    opacity: 0.9;
  }

  .btn-delete {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
    color: var(--color-text);
  }

  .btn-delete:hover {
    background-color: #fee2e2;
    border-color: #f87171;
  }

  /* Composer Modal */
  .composer-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .composer-modal {
    background-color: var(--color-surface);
    border-radius: var(--radius-lg);
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  .composer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  .composer-header h3 {
    margin: 0;
    font-size: 1rem;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--color-text-muted);
  }

  .close-btn:hover {
    color: var(--color-text);
  }

  .composer-body {
    flex: 1;
    padding: var(--space-4);
    overflow-y: auto;
  }

  .composer-body textarea {
    width: 100%;
    height: 200px;
    padding: var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    font-family: inherit;
    font-size: 0.875rem;
    resize: none;
  }

  .composer-footer {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    padding: var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  .btn-send,
  .btn-cancel {
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 0.875rem;
  }

  .btn-send {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-send:hover {
    opacity: 0.9;
  }

  .btn-cancel {
    background-color: var(--color-surface);
    border: 1px solid var(--color-border);
  }

  .btn-cancel:hover {
    background-color: var(--color-surface-raised);
  }
</style>

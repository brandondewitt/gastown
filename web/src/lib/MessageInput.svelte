<script lang="ts">
  let {
    agentAddress,
    disabled = false,
    placeholder = 'Type a message...',
    onSent = () => {}
  } = $props<{
    agentAddress: string;
    disabled?: boolean;
    placeholder?: string;
    onSent?: () => void;
  }>();

  let message = $state('');
  let sending = $state(false);
  let error = $state<string | null>(null);

  async function sendMessage() {
    if (!message.trim() || sending || disabled) return;

    sending = true;
    error = null;

    try {
      const res = await fetch(`/api/v1/agents/${agentAddress}/message`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: message.trim() }),
      });

      const data = await res.json();

      if (!res.ok || !data.success) {
        throw new Error(data.error?.message || 'Failed to send message');
      }

      message = '';
      onSent();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to send message';
      // Clear error after 3 seconds
      setTimeout(() => error = null, 3000);
    } finally {
      sending = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }
</script>

<div class="message-input-container">
  {#if error}
    <div class="error-banner">
      <span>⚠️ {error}</span>
    </div>
  {/if}
  <div class="message-input">
    <input
      type="text"
      bind:value={message}
      onkeydown={handleKeydown}
      placeholder={disabled ? 'Agent is not running...' : placeholder}
      disabled={disabled || sending}
      class:has-error={error}
    />
    <button
      onclick={sendMessage}
      disabled={disabled || sending || !message.trim()}
      class="send-button"
    >
      {#if sending}
        <span class="sending-icon">...</span>
      {:else}
        <span class="send-icon">▶</span>
      {/if}
    </button>
  </div>
</div>

<style>
  .message-input-container {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .message-input {
    display: flex;
    gap: var(--space-2);
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: var(--space-2);
  }

  .message-input:focus-within {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  input {
    flex: 1;
    background: none;
    border: none;
    color: var(--color-text);
    font-size: 0.875rem;
    padding: var(--space-2);
    outline: none;
  }

  input::placeholder {
    color: var(--color-text-muted);
  }

  input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  input.has-error {
    color: var(--color-error);
  }

  .send-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    height: 2.5rem;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all var(--transition-fast);
    flex-shrink: 0;
  }

  .send-button:hover:not(:disabled) {
    opacity: 0.9;
  }

  .send-button:disabled {
    background: var(--color-text-muted);
    cursor: not-allowed;
    opacity: 0.5;
  }

  .send-icon {
    font-size: 0.875rem;
  }

  .sending-icon {
    font-size: 1rem;
    animation: pulse 1s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }

  .error-banner {
    padding: var(--space-2) var(--space-3);
    background: rgba(248, 81, 73, 0.1);
    border: 1px solid var(--color-error);
    border-radius: var(--radius-sm);
    color: var(--color-error);
    font-size: 0.75rem;
  }
</style>

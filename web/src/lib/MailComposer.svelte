<script lang="ts">
  interface MailComposerProps {
    onSent?: () => void;
  }

  let { onSent }: MailComposerProps = $props();

  let to = $state('');
  let subject = $state('');
  let body = $state('');
  let loading = $state(false);
  let error: string | null = $state(null);
  let success = $state(false);

  async function handleSend() {
    // Validation
    if (!to.trim()) {
      error = 'Recipient address is required';
      return;
    }
    if (!subject.trim()) {
      error = 'Subject is required';
      return;
    }
    if (!body.trim()) {
      error = 'Message body is required';
      return;
    }

    loading = true;
    error = null;
    success = false;

    try {
      const response = await fetch('/api/v1/mail', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          to: to.trim(),
          subject: subject.trim(),
          body: body.trim(),
        }),
      });

      const data = await response.json();

      if (data.success) {
        success = true;
        // Reset form
        to = '';
        subject = '';
        body = '';
        // Clear success message after 3 seconds
        setTimeout(() => {
          success = false;
        }, 3000);
        // Call callback if provided
        onSent?.();
      } else {
        error = data.error?.message || 'Failed to send mail';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to send mail';
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      handleSend();
    }
  }
</script>

<div class="mail-composer card">
  <h2 class="card-header">Compose Mail</h2>

  {#if success}
    <div class="success-message">
      ✓ Mail sent successfully
    </div>
  {/if}

  {#if error}
    <div class="error-message">
      {error}
    </div>
  {/if}

  <form onsubmit={(e) => { e.preventDefault(); handleSend(); }}>
    <div class="form-group">
      <label for="to" class="form-label">To</label>
      <input
        id="to"
        type="text"
        class="form-input"
        placeholder="e.g., mayor/, crew/slit, deacon/"
        bind:value={to}
        disabled={loading}
      />
      <p class="form-hint">Recipient address (required)</p>
    </div>

    <div class="form-group">
      <label for="subject" class="form-label">Subject</label>
      <input
        id="subject"
        type="text"
        class="form-input"
        placeholder="Brief subject line"
        bind:value={subject}
        disabled={loading}
      />
      <p class="form-hint">Subject line (required)</p>
    </div>

    <div class="form-group">
      <label for="body" class="form-label">Message</label>
      <textarea
        id="body"
        class="form-textarea"
        placeholder="Your message here..."
        bind:value={body}
        disabled={loading}
        onkeydown={handleKeydown}
      ></textarea>
      <p class="form-hint">Message body (required) • Ctrl+Enter to send</p>
    </div>

    <div class="form-actions">
      <button
        type="submit"
        class="btn btn-primary"
        disabled={loading}
      >
        {#if loading}
          <span class="spinner-small"></span>
          Sending...
        {:else}
          Send Mail
        {/if}
      </button>
    </div>
  </form>
</div>

<style>
  .mail-composer {
    max-width: 600px;
  }

  .success-message {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
    margin-bottom: var(--space-4);
    background-color: rgba(34, 197, 94, 0.15);
    border: 1px solid var(--color-success);
    border-radius: var(--radius-md);
    color: var(--color-success);
    font-size: 0.875rem;
    font-weight: 500;
  }

  .error-message {
    padding: var(--space-3) var(--space-4);
    margin-bottom: var(--space-4);
    background-color: rgba(239, 68, 68, 0.15);
    border: 1px solid var(--color-error);
    border-radius: var(--radius-md);
    color: var(--color-error);
    font-size: 0.875rem;
    font-weight: 500;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .form-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text);
  }

  .form-input,
  .form-textarea {
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-surface-raised);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    color: var(--color-text);
    font-family: var(--font-sans);
    font-size: 0.875rem;
    transition: border-color var(--transition-fast), background-color var(--transition-fast);
  }

  .form-input:focus,
  .form-textarea:focus {
    outline: none;
    border-color: var(--color-primary);
    background-color: var(--color-surface);
  }

  .form-input:disabled,
  .form-textarea:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .form-textarea {
    min-height: 120px;
    resize: vertical;
    font-family: var(--font-mono);
  }

  .form-hint {
    font-size: 0.75rem;
    color: var(--color-text-muted);
    margin-top: -var(--space-1);
  }

  .form-actions {
    display: flex;
    gap: var(--space-3);
    padding-top: var(--space-2);
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color var(--transition-fast), opacity var(--transition-fast);
  }

  .btn-primary {
    background-color: var(--color-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .spinner-small {
    display: inline-block;
    width: 0.875rem;
    height: 0.875rem;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>

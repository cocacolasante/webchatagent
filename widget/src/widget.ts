// Blueprint Chat Widget — Vanilla TypeScript
// Compiles to a single self-contained IIFE bundle
import widgetCSS from './styles/widget.css';

interface WidgetConfig {
  tenantId: string;
  botName: string;
  avatarUrl?: string;
  primaryColor: string;
  accentColor: string;
  position: 'bottom-right' | 'bottom-left';
  greeting: string;
  placeholderText: string;
  features: {
    leadCapture: boolean;
    booking: boolean;
    schedulerType?: string;
  };
  leadForm: {
    fields: string[];
    requiredFields: string[];
    triggerMessage: string;
  };
}

interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
  isStreaming?: boolean;
}

type WidgetState = 'closed' | 'open' | 'typing' | 'lead_form' | 'booking_slots' | 'booking_confirm' | 'error';

interface TimeSlot {
  startTime: string;
  endTime: string;
  available: boolean;
}

(function BlueprintChat() {
  'use strict';

  const scriptEl = document.currentScript as HTMLScriptElement | null;
  const TENANT_ID = scriptEl?.getAttribute('data-tenant-id') || '';
  const POSITION = (scriptEl?.getAttribute('data-position') || 'bottom-right') as 'bottom-right' | 'bottom-left';
  const API_BASE = scriptEl?.getAttribute('data-api-base') || 'https://chat.blueprintautomation.tech';

  if (!TENANT_ID) {
    console.warn('[Blueprint Chat] Missing data-tenant-id attribute');
    return;
  }

  let config: WidgetConfig | null = null;
  let sessionId: string | null = null;
  let state: WidgetState = 'closed';
  let messages: Message[] = [];
  let currentAbortController: AbortController | null = null;

  // DOM references
  let bubble: HTMLElement | null = null;
  let panel: HTMLElement | null = null;
  let messagesContainer: HTMLElement | null = null;
  let inputEl: HTMLTextAreaElement | null = null;
  let sendBtn: HTMLButtonElement | null = null;
  let typingIndicator: HTMLElement | null = null;

  // ── CSS injection ────────────────────────────────────────────────────────────

  function injectCSS(primaryColor: string): void {
    const style = document.createElement('style');
    style.textContent = widgetCSS;
    style.setAttribute('data-bp-chat', '1');
    document.head.appendChild(style);

    // Inject CSS custom property for primary color
    const colorStyle = document.createElement('style');
    colorStyle.textContent = `:root { --bp-primary: ${primaryColor}; }`;
    document.head.appendChild(colorStyle);
  }

  // ── API calls ────────────────────────────────────────────────────────────────

  async function fetchConfig(): Promise<WidgetConfig> {
    const resp = await fetch(`${API_BASE}/widget-config?tid=${TENANT_ID}`);
    if (!resp.ok) throw new Error(`Config fetch failed: ${resp.status}`);
    return resp.json();
  }

  async function sendMessage(message: string): Promise<void> {
    if (!config) return;
    setState('typing');
    appendMessage('user', message);
    showTypingIndicator();

    currentAbortController = new AbortController();

    try {
      const resp = await fetch(`${API_BASE}/api/chat/stream?tid=${TENANT_ID}${sessionId ? `&sid=${sessionId}` : ''}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message,
          sessionId: sessionId || undefined,
          sourceUrl: window.location.href,
        }),
        signal: currentAbortController.signal,
      });

      if (!resp.ok) {
        throw new Error(`Chat request failed: ${resp.status}`);
      }

      const reader = resp.body!.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let assistantMsgEl: HTMLElement | null = null;
      let assistantContent = '';

      hideTypingIndicator();
      assistantMsgEl = createMessageElement('assistant', '');
      messagesContainer?.appendChild(assistantMsgEl);

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            continue; // event type handled below
          }
          if (line.startsWith('data: ')) {
            const eventType = getEventType(lines, line);
            const dataStr = line.slice(6).trim();
            if (!dataStr) continue;

            try {
              const data = JSON.parse(dataStr);
              handleSSEData(eventType, data, assistantMsgEl, (text: string) => {
                assistantContent += text;
                updateMessageContent(assistantMsgEl!, assistantContent);
                scrollToBottom();
              });
            } catch {
              // ignore parse errors
            }
          }
        }
      }

      setState('open');
      scrollToBottom();
    } catch (err: unknown) {
      hideTypingIndicator();
      setState('open');
      if (err instanceof Error && err.name !== 'AbortError') {
        appendMessage('system', 'Something went wrong. Please try again.');
        console.error('[Blueprint Chat] Stream error:', err);
      }
    }
  }

  function getEventType(lines: string[], dataLine: string): string {
    const dataIdx = lines.indexOf(dataLine);
    for (let i = dataIdx - 1; i >= 0; i--) {
      if (lines[i].startsWith('event: ')) {
        return lines[i].slice(7).trim();
      }
    }
    return 'message';
  }

  function handleSSEData(
    eventType: string,
    data: Record<string, unknown>,
    msgEl: HTMLElement | null,
    onChunk: (text: string) => void
  ): void {
    switch (eventType) {
      case 'chunk':
        onChunk((data.text as string) || '');
        break;
      case 'done':
        if (data.sessionId) sessionId = data.sessionId as string;
        if (data.leadPromptSuggested && config?.features.leadCapture && !sessionId) {
          setTimeout(() => showLeadForm(), 800);
        }
        if (data.bookingPromptSuggested && config?.features.booking) {
          setTimeout(() => appendBookingPrompt(), 800);
        }
        break;
      case 'error':
        appendMessage('system', (data.message as string) || 'An error occurred.');
        break;
      case 'system':
        if (data.type === 'lead_form') {
          setTimeout(() => showLeadForm(), 500);
        }
        if (data.type === 'booking_prompt') {
          setTimeout(() => appendBookingPrompt(), 500);
        }
        break;
    }
  }

  // ── UI helpers ────────────────────────────────────────────────────────────────

  function setState(newState: WidgetState): void {
    state = newState;
  }

  function scrollToBottom(): void {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  }

  function appendMessage(role: Message['role'], content: string): void {
    messages.push({ role, content });
    const el = createMessageElement(role, content);
    messagesContainer?.appendChild(el);
    scrollToBottom();
  }

  function createMessageElement(role: string, content: string): HTMLElement {
    const wrapper = document.createElement('div');
    wrapper.className = `bp-chat-msg bp-chat-msg--${role}`;

    const bubble = document.createElement('div');
    bubble.className = 'bp-chat-msg-bubble';
    bubble.textContent = content;

    wrapper.appendChild(bubble);
    return wrapper;
  }

  function updateMessageContent(el: HTMLElement, content: string): void {
    const bubble = el.querySelector('.bp-chat-msg-bubble');
    if (bubble) bubble.textContent = content;
  }

  function showTypingIndicator(): void {
    typingIndicator = document.createElement('div');
    typingIndicator.className = 'bp-chat-typing';
    typingIndicator.innerHTML = `
      <div class="bp-chat-typing-dot"></div>
      <div class="bp-chat-typing-dot"></div>
      <div class="bp-chat-typing-dot"></div>
    `;
    messagesContainer?.appendChild(typingIndicator);
    scrollToBottom();
  }

  function hideTypingIndicator(): void {
    typingIndicator?.remove();
    typingIndicator = null;
  }

  function showLeadForm(): void {
    if (!config) return;
    const form = document.createElement('div');
    form.className = 'bp-chat-lead-form';
    form.innerHTML = `
      <h4>Stay in touch</h4>
      <p>${config.leadForm.triggerMessage || "I'd love to help — could I get your contact info?"}</p>
      ${config.leadForm.fields.includes('name') ? `<input class="bp-chat-lead-field" name="name" placeholder="Your name${config.leadForm.requiredFields.includes('name') ? ' *' : ''}" />` : ''}
      ${config.leadForm.fields.includes('email') ? `<input class="bp-chat-lead-field" type="email" name="email" placeholder="Email address${config.leadForm.requiredFields.includes('email') ? ' *' : ''}" />` : ''}
      ${config.leadForm.fields.includes('phone') ? `<input class="bp-chat-lead-field" type="tel" name="phone" placeholder="Phone number" />` : ''}
      <button class="bp-chat-lead-submit">Send my details</button>
    `;

    const submitBtn = form.querySelector('.bp-chat-lead-submit') as HTMLButtonElement;
    submitBtn.addEventListener('click', () => submitLeadForm(form));

    messagesContainer?.appendChild(form);
    scrollToBottom();
  }

  async function submitLeadForm(form: HTMLElement): Promise<void> {
    const name = (form.querySelector('[name="name"]') as HTMLInputElement)?.value || '';
    const email = (form.querySelector('[name="email"]') as HTMLInputElement)?.value || '';
    const phone = (form.querySelector('[name="phone"]') as HTMLInputElement)?.value || '';

    if (!email) {
      alert('Please enter your email address.');
      return;
    }

    const submitBtn = form.querySelector('.bp-chat-lead-submit') as HTMLButtonElement;
    submitBtn.disabled = true;
    submitBtn.textContent = 'Sending...';

    const nameParts = name.trim().split(' ');
    const firstName = nameParts[0] || 'Friend';
    const lastName = nameParts.slice(1).join(' ');

    try {
      const resp = await fetch(`${API_BASE}/api/leads?tid=${TENANT_ID}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          firstName,
          lastName,
          email,
          phone,
          sessionId,
          sourceUrl: window.location.href,
        }),
      });

      if (resp.ok) {
        form.innerHTML = `<div class="bp-chat-confirm-msg">✓ Thanks ${firstName}! We'll be in touch soon.</div>`;
        setTimeout(() => scrollToBottom(), 100);
      } else {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Try again';
      }
    } catch {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Try again';
    }
  }

  function appendBookingPrompt(): void {
    const promptEl = document.createElement('div');
    promptEl.className = 'bp-chat-msg bp-chat-msg--assistant';
    promptEl.innerHTML = `
      <div class="bp-chat-msg-bubble">
        Would you like to see available times to connect?
        <br/><br/>
        <button class="bp-chat-slot-btn" style="margin-top:4px; width:100%">📅 Show Available Times</button>
      </div>
    `;
    promptEl.querySelector('button')?.addEventListener('click', () => loadBookingSlots());
    messagesContainer?.appendChild(promptEl);
    scrollToBottom();
  }

  async function loadBookingSlots(): Promise<void> {
    const slotsContainer = document.createElement('div');
    slotsContainer.className = 'bp-chat-booking-slots';
    slotsContainer.innerHTML = '<h4>Available Times</h4><p style="font-size:13px;color:#666">Loading...</p>';
    messagesContainer?.appendChild(slotsContainer);
    scrollToBottom();

    try {
      const resp = await fetch(`${API_BASE}/api/booking/availability?tid=${TENANT_ID}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });

      const data = await resp.json();
      const slots: TimeSlot[] = (data.slots || []).slice(0, 8);

      if (slots.length === 0) {
        slotsContainer.innerHTML = '<h4>Available Times</h4><p style="color:#666;font-size:13px">No slots available right now. Please contact us directly.</p>';
        return;
      }

      const slotList = document.createElement('div');
      slotList.className = 'bp-chat-slot-list';

      slots.forEach(slot => {
        const btn = document.createElement('button');
        btn.className = 'bp-chat-slot-btn';
        const start = new Date(slot.startTime);
        btn.textContent = start.toLocaleString('en-US', {
          weekday: 'short', month: 'short', day: 'numeric',
          hour: 'numeric', minute: '2-digit', hour12: true,
        });
        btn.addEventListener('click', () => selectSlot(slot, slotsContainer));
        slotList.appendChild(btn);
      });

      slotsContainer.innerHTML = '<h4>Pick a time</h4>';
      slotsContainer.appendChild(slotList);
      scrollToBottom();
    } catch {
      slotsContainer.innerHTML = '<div class="bp-chat-error-msg">Could not load available times. Please contact us directly.</div>';
    }
  }

  function selectSlot(slot: TimeSlot, container: HTMLElement): void {
    container.innerHTML = `
      <h4>Confirm your booking</h4>
      <p style="font-size:13px; color:#444">${new Date(slot.startTime).toLocaleString('en-US', { weekday:'long', month:'long', day:'numeric', hour:'numeric', minute:'2-digit', hour12:true })}</p>
      <input class="bp-chat-lead-field" name="name" placeholder="Your name *" />
      <input class="bp-chat-lead-field" type="email" name="email" placeholder="Email address *" />
      <input class="bp-chat-lead-field" name="notes" placeholder="Anything we should know? (optional)" />
      <button class="bp-chat-lead-submit">Confirm Booking</button>
    `;

    container.querySelector('.bp-chat-lead-submit')?.addEventListener('click', async () => {
      const name = (container.querySelector('[name="name"]') as HTMLInputElement)?.value || '';
      const email = (container.querySelector('[name="email"]') as HTMLInputElement)?.value || '';
      const notes = (container.querySelector('[name="notes"]') as HTMLInputElement)?.value || '';

      if (!name || !email) { alert('Name and email are required.'); return; }

      const btn = container.querySelector('.bp-chat-lead-submit') as HTMLButtonElement;
      btn.disabled = true;
      btn.textContent = 'Booking...';

      try {
        const resp = await fetch(`${API_BASE}/api/booking/create?tid=${TENANT_ID}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            slotStartTime: slot.startTime,
            attendeeName: name,
            attendeeEmail: email,
            notes,
          }),
        });

        if (resp.ok) {
          const appt = await resp.json();
          container.innerHTML = `
            <div class="bp-chat-confirm-msg">
              ✓ <strong>Booking confirmed!</strong><br/>
              ${new Date(appt.startTime || slot.startTime).toLocaleString('en-US', { weekday:'long', month:'long', day:'numeric', hour:'numeric', minute:'2-digit', hour12:true })}
              ${appt.meetingUrl ? `<br/><a href="${appt.meetingUrl}" target="_blank" style="color:#2e7d32">Join meeting link</a>` : ''}
              ${appt.confirmUrl ? `<br/><a href="${appt.confirmUrl}" target="_blank" style="color:#2e7d32; font-size:12px">View booking details</a>` : ''}
            </div>`;
        } else {
          btn.disabled = false;
          btn.textContent = 'Try again';
        }
      } catch {
        btn.disabled = false;
        btn.textContent = 'Try again';
      }
      scrollToBottom();
    });
    scrollToBottom();
  }

  // ── Widget render ─────────────────────────────────────────────────────────────

  function renderBubble(cfg: WidgetConfig): void {
    bubble = document.createElement('button');
    bubble.className = `bp-chat-bubble bp-chat-bubble--${cfg.position}`;
    bubble.setAttribute('aria-label', `Chat with ${cfg.botName}`);
    bubble.innerHTML = `<svg viewBox="0 0 24 24"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/></svg>`;
    bubble.addEventListener('click', togglePanel);
    document.body.appendChild(bubble);
  }

  function renderPanel(cfg: WidgetConfig): void {
    panel = document.createElement('div');
    panel.className = `bp-chat-panel bp-chat-panel--${cfg.position} bp-chat-panel--hidden`;
    panel.innerHTML = `
      <div class="bp-chat-header">
        <div class="bp-chat-header-avatar">
          ${cfg.avatarUrl ? `<img src="${cfg.avatarUrl}" alt="${cfg.botName}" />` : cfg.botName.charAt(0)}
        </div>
        <div class="bp-chat-header-info">
          <div class="bp-chat-header-name">${cfg.botName}</div>
          <div class="bp-chat-header-status">Online · Usually replies instantly</div>
        </div>
        <button class="bp-chat-header-close" aria-label="Close chat">
          <svg viewBox="0 0 24 24"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
        </button>
      </div>
      <div class="bp-chat-messages"></div>
      <div class="bp-chat-input-area">
        <textarea class="bp-chat-input" placeholder="${cfg.placeholderText || 'Type a message...'}" rows="1"></textarea>
        <button class="bp-chat-send-btn" aria-label="Send">
          <svg viewBox="0 0 24 24"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>
        </button>
      </div>
      <div class="bp-chat-powered"><a href="https://blueprintautomation.tech" target="_blank" rel="noopener">Powered by Blueprint Chat</a></div>
    `;

    messagesContainer = panel.querySelector('.bp-chat-messages');
    inputEl = panel.querySelector('.bp-chat-input');
    sendBtn = panel.querySelector('.bp-chat-send-btn');

    panel.querySelector('.bp-chat-header-close')?.addEventListener('click', closePanel);
    sendBtn?.addEventListener('click', handleSend);
    inputEl?.addEventListener('keydown', (e: KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    });
    inputEl?.addEventListener('input', () => autoResize(inputEl!));

    document.body.appendChild(panel);
  }

  function autoResize(el: HTMLTextAreaElement): void {
    el.style.height = 'auto';
    el.style.height = Math.min(el.scrollHeight, 100) + 'px';
  }

  function handleSend(): void {
    if (!inputEl || !sendBtn || state === 'typing') return;
    const msg = inputEl.value.trim();
    if (!msg) return;
    inputEl.value = '';
    inputEl.style.height = 'auto';
    sendMessage(msg);
  }

  function togglePanel(): void {
    if (!panel) return;
    if (panel.classList.contains('bp-chat-panel--hidden')) {
      openPanel();
    } else {
      closePanel();
    }
  }

  function openPanel(): void {
    panel?.classList.remove('bp-chat-panel--hidden');
    bubble?.classList.remove('bp-chat-bubble--pulse');
    setState('open');

    if (messages.length === 0 && config) {
      appendMessage('assistant', config.greeting);
    }

    inputEl?.focus();
  }

  function closePanel(): void {
    panel?.classList.add('bp-chat-panel--hidden');
    setState('closed');
    currentAbortController?.abort();
  }

  // ── Init ──────────────────────────────────────────────────────────────────────

  async function init(): Promise<void> {
    try {
      config = await fetchConfig();
    } catch (err) {
      console.error('[Blueprint Chat] Failed to load config:', err);
      return;
    }

    injectCSS(config.primaryColor);
    renderBubble(config);
    renderPanel(config);
  }

  // Boot when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();

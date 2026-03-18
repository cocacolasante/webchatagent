"use strict";(()=>{var A=`/* Blueprint Chat Widget \u2014 all classes prefixed with bp-chat- to avoid conflicts */

.bp-chat-bubble {
  position: fixed;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bp-primary, #6C63FF);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 16px rgba(0,0,0,0.18);
  z-index: 2147483640;
  border: none;
  outline: none;
  transition: transform 0.2s, box-shadow 0.2s;
}

.bp-chat-bubble:hover {
  transform: scale(1.08);
  box-shadow: 0 6px 20px rgba(0,0,0,0.22);
}

.bp-chat-bubble svg {
  width: 26px;
  height: 26px;
  fill: #ffffff;
}

.bp-chat-bubble--bottom-right { bottom: 24px; right: 24px; }
.bp-chat-bubble--bottom-left  { bottom: 24px; left: 24px; }

.bp-chat-bubble--pulse::after {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 2px solid var(--bp-primary, #6C63FF);
  animation: bp-pulse 2s ease-out infinite;
}

@keyframes bp-pulse {
  0%   { transform: scale(1); opacity: 0.8; }
  100% { transform: scale(1.5); opacity: 0; }
}

.bp-chat-panel {
  position: fixed;
  width: 380px;
  max-width: calc(100vw - 32px);
  height: 600px;
  max-height: calc(100vh - 100px);
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0,0,0,0.16);
  display: flex;
  flex-direction: column;
  z-index: 2147483641;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 14px;
  color: #1a1a1a;
}

.bp-chat-panel--bottom-right { bottom: 92px; right: 24px; }
.bp-chat-panel--bottom-left  { bottom: 92px; left: 24px; }

.bp-chat-panel--hidden {
  display: none;
}

.bp-chat-header {
  background: var(--bp-primary, #6C63FF);
  color: #ffffff;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.bp-chat-header-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255,255,255,0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  flex-shrink: 0;
}

.bp-chat-header-avatar img {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}

.bp-chat-header-info { flex: 1; }
.bp-chat-header-name { font-weight: 600; font-size: 15px; }
.bp-chat-header-status { font-size: 12px; opacity: 0.85; margin-top: 1px; }

.bp-chat-header-close {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  padding: 4px;
  opacity: 0.8;
  border-radius: 4px;
  display: flex;
  align-items: center;
}
.bp-chat-header-close:hover { opacity: 1; background: rgba(255,255,255,0.15); }
.bp-chat-header-close svg { width: 18px; height: 18px; fill: currentColor; }

.bp-chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  scroll-behavior: smooth;
}

.bp-chat-messages::-webkit-scrollbar { width: 4px; }
.bp-chat-messages::-webkit-scrollbar-track { background: transparent; }
.bp-chat-messages::-webkit-scrollbar-thumb { background: #e0e0e0; border-radius: 4px; }

.bp-chat-msg {
  display: flex;
  gap: 8px;
  max-width: 88%;
}

.bp-chat-msg--user { align-self: flex-end; flex-direction: row-reverse; }
.bp-chat-msg--assistant { align-self: flex-start; }

.bp-chat-msg-bubble {
  padding: 10px 14px;
  border-radius: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.bp-chat-msg--user .bp-chat-msg-bubble {
  background: var(--bp-primary, #6C63FF);
  color: #fff;
  border-bottom-right-radius: 4px;
}

.bp-chat-msg--assistant .bp-chat-msg-bubble {
  background: #f0f0f0;
  color: #1a1a1a;
  border-bottom-left-radius: 4px;
}

.bp-chat-typing {
  align-self: flex-start;
  display: flex;
  gap: 4px;
  padding: 12px 16px;
  background: #f0f0f0;
  border-radius: 14px;
  border-bottom-left-radius: 4px;
}

.bp-chat-typing-dot {
  width: 6px;
  height: 6px;
  background: #999;
  border-radius: 50%;
  animation: bp-typing 1.2s ease-in-out infinite;
}

.bp-chat-typing-dot:nth-child(2) { animation-delay: 0.2s; }
.bp-chat-typing-dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes bp-typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30%            { transform: translateY(-4px); opacity: 1; }
}

.bp-chat-input-area {
  padding: 12px 16px;
  border-top: 1px solid #ebebeb;
  display: flex;
  gap: 8px;
  align-items: flex-end;
  flex-shrink: 0;
  background: #fff;
}

.bp-chat-input {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 20px;
  padding: 10px 16px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  outline: none;
  max-height: 100px;
  line-height: 1.4;
  color: #1a1a1a;
  background: #fafafa;
  transition: border-color 0.15s;
}

.bp-chat-input:focus { border-color: var(--bp-primary, #6C63FF); background: #fff; }

.bp-chat-send-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--bp-primary, #6C63FF);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: opacity 0.15s;
}
.bp-chat-send-btn:hover { opacity: 0.88; }
.bp-chat-send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.bp-chat-send-btn svg { width: 16px; height: 16px; fill: #fff; }

/* Lead form */
.bp-chat-lead-form {
  background: #f8f8ff;
  border: 1px solid #e8e6ff;
  border-radius: 12px;
  padding: 16px;
  margin: 4px 0;
}

.bp-chat-lead-form h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
}

.bp-chat-lead-form p {
  margin: 0 0 12px;
  font-size: 13px;
  color: #666;
}

.bp-chat-lead-field {
  width: 100%;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 9px 12px;
  font-size: 13px;
  font-family: inherit;
  margin-bottom: 8px;
  outline: none;
  box-sizing: border-box;
  background: #fff;
}
.bp-chat-lead-field:focus { border-color: var(--bp-primary, #6C63FF); }

.bp-chat-lead-submit {
  width: 100%;
  background: var(--bp-primary, #6C63FF);
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
  font-family: inherit;
}
.bp-chat-lead-submit:hover { opacity: 0.88; }
.bp-chat-lead-submit:disabled { opacity: 0.5; cursor: not-allowed; }

/* Booking slots */
.bp-chat-booking-slots {
  background: #f8fff8;
  border: 1px solid #e0ffe0;
  border-radius: 12px;
  padding: 16px;
  margin: 4px 0;
}

.bp-chat-booking-slots h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
}

.bp-chat-slot-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 10px;
  max-height: 200px;
  overflow-y: auto;
}

.bp-chat-slot-btn {
  background: #fff;
  border: 1px solid #c8e6c9;
  border-radius: 8px;
  padding: 9px 14px;
  font-size: 13px;
  cursor: pointer;
  text-align: left;
  transition: background 0.12s, border-color 0.12s;
  font-family: inherit;
}
.bp-chat-slot-btn:hover { background: var(--bp-primary, #6C63FF); color: #fff; border-color: transparent; }

.bp-chat-confirm-msg {
  background: #f0fff0;
  border: 1px solid #b2dfdb;
  border-radius: 12px;
  padding: 14px 16px;
  font-size: 13px;
  color: #2e7d32;
}

.bp-chat-error-msg {
  background: #fff5f5;
  border: 1px solid #ffcdd2;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  color: #c62828;
}

.bp-chat-powered {
  text-align: center;
  font-size: 11px;
  color: #bbb;
  padding: 6px 0 10px;
  flex-shrink: 0;
}

.bp-chat-powered a { color: #bbb; text-decoration: none; }
.bp-chat-powered a:hover { color: #999; }

/* Mobile */
@media (max-width: 480px) {
  .bp-chat-panel {
    width: 100vw;
    max-width: 100vw;
    height: 100vh;
    max-height: 100vh;
    bottom: 0 !important;
    right: 0 !important;
    left: 0 !important;
    border-radius: 0;
  }
}
`;(function(){"use strict";let h=document.currentScript||document.querySelector("script[data-tenant-id]"),v=(h==null?void 0:h.getAttribute("data-tenant-id"))||"",te=(h==null?void 0:h.getAttribute("data-position"))||"bottom-right",k=(h==null?void 0:h.getAttribute("data-api-base"))||"https://chat.blueprintautomation.tech";if(!v){console.warn("[Blueprint Chat] Missing data-tenant-id attribute");return}let o=null,w=null,E="closed",M=[],T=null,u=null,l=null,a=null,d=null,S=null,g=null;function O(e){let n=document.createElement("style");n.textContent=A,n.setAttribute("data-bp-chat","1"),document.head.appendChild(n);let t=document.createElement("style");t.textContent=`:root { --bp-primary: ${e}; }`,document.head.appendChild(t)}async function U(){let e=await fetch(`${k}/widget-config?tid=${v}`);if(!e.ok)throw new Error(`Config fetch failed: ${e.status}`);return e.json()}async function W(e){if(o){L("typing"),C("user",e),J(),T=new AbortController;try{let n=await fetch(`${k}/api/chat/stream?tid=${v}${w?`&sid=${w}`:""}`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({message:e,sessionId:w||void 0,sourceUrl:window.location.href}),signal:T.signal});if(!n.ok)throw new Error(`Chat request failed: ${n.status}`);let t=n.body.getReader(),i=new TextDecoder,r="",c=null,s="";for($(),c=H("assistant",""),a==null||a.appendChild(c);;){let{done:x,value:y}=await t.read();if(x)break;r+=i.decode(y,{stream:!0});let m=r.split(`
`);r=m.pop()||"";for(let b of m)if(!b.startsWith("event: ")&&b.startsWith("data: ")){let f=j(m,b),q=b.slice(6).trim();if(!q)continue;try{let I=JSON.parse(q);_(f,I,c,ee=>{s+=ee,D(c,s),p()})}catch(I){}}}L("open"),p()}catch(n){$(),L("open"),n instanceof Error&&n.name!=="AbortError"&&(C("system","Something went wrong. Please try again."),console.error("[Blueprint Chat] Stream error:",n))}}}function j(e,n){let t=e.indexOf(n);for(let i=t-1;i>=0;i--)if(e[i].startsWith("event: "))return e[i].slice(7).trim();return"message"}function _(e,n,t,i){switch(e){case"chunk":i(n.text||"");break;case"done":n.sessionId&&(w=n.sessionId),n.leadPromptSuggested&&(o!=null&&o.features.leadCapture)&&!w&&setTimeout(()=>F(),800),n.bookingPromptSuggested&&(o!=null&&o.features.booking)&&setTimeout(()=>N(),800);break;case"error":C("system",n.message||"An error occurred.");break;case"system":n.type==="lead_form"&&setTimeout(()=>F(),500),n.type==="booking_prompt"&&setTimeout(()=>N(),500);break}}function L(e){E=e}function p(){a&&(a.scrollTop=a.scrollHeight)}function C(e,n){M.push({role:e,content:n});let t=H(e,n);a==null||a.appendChild(t),p()}function H(e,n){let t=document.createElement("div");t.className=`bp-chat-msg bp-chat-msg--${e}`;let i=document.createElement("div");return i.className="bp-chat-msg-bubble",i.textContent=n,t.appendChild(i),t}function D(e,n){let t=e.querySelector(".bp-chat-msg-bubble");t&&(t.textContent=n)}function J(){g=document.createElement("div"),g.className="bp-chat-typing",g.innerHTML=`
      <div class="bp-chat-typing-dot"></div>
      <div class="bp-chat-typing-dot"></div>
      <div class="bp-chat-typing-dot"></div>
    `,a==null||a.appendChild(g),p()}function $(){g==null||g.remove(),g=null}function F(){if(!o)return;let e=document.createElement("div");e.className="bp-chat-lead-form",e.innerHTML=`
      <h4>Stay in touch</h4>
      <p>${o.leadForm.triggerMessage||"I'd love to help \u2014 could I get your contact info?"}</p>
      ${o.leadForm.fields.includes("name")?`<input class="bp-chat-lead-field" name="name" placeholder="Your name${o.leadForm.requiredFields.includes("name")?" *":""}" />`:""}
      ${o.leadForm.fields.includes("email")?`<input class="bp-chat-lead-field" type="email" name="email" placeholder="Email address${o.leadForm.requiredFields.includes("email")?" *":""}" />`:""}
      ${o.leadForm.fields.includes("phone")?'<input class="bp-chat-lead-field" type="tel" name="phone" placeholder="Phone number" />':""}
      <button class="bp-chat-lead-submit">Send my details</button>
    `,e.querySelector(".bp-chat-lead-submit").addEventListener("click",()=>R(e)),a==null||a.appendChild(e),p()}async function R(e){var y,m,b;let n=((y=e.querySelector('[name="name"]'))==null?void 0:y.value)||"",t=((m=e.querySelector('[name="email"]'))==null?void 0:m.value)||"",i=((b=e.querySelector('[name="phone"]'))==null?void 0:b.value)||"";if(!t){alert("Please enter your email address.");return}let r=e.querySelector(".bp-chat-lead-submit");r.disabled=!0,r.textContent="Sending...";let c=n.trim().split(" "),s=c[0]||"Friend",x=c.slice(1).join(" ");try{(await fetch(`${k}/api/leads?tid=${v}`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({firstName:s,lastName:x,email:t,phone:i,sessionId:w,sourceUrl:window.location.href})})).ok?(e.innerHTML=`<div class="bp-chat-confirm-msg">\u2713 Thanks ${s}! We'll be in touch soon.</div>`,setTimeout(()=>p(),100)):(r.disabled=!1,r.textContent="Try again")}catch(f){r.disabled=!1,r.textContent="Try again"}}function N(){var n;let e=document.createElement("div");e.className="bp-chat-msg bp-chat-msg--assistant",e.innerHTML=`
      <div class="bp-chat-msg-bubble">
        Would you like to see available times to connect?
        <br/><br/>
        <button class="bp-chat-slot-btn" style="margin-top:4px; width:100%">\u{1F4C5} Show Available Times</button>
      </div>
    `,(n=e.querySelector("button"))==null||n.addEventListener("click",()=>Y()),a==null||a.appendChild(e),p()}async function Y(){let e=document.createElement("div");e.className="bp-chat-booking-slots",e.innerHTML='<h4>Available Times</h4><p style="font-size:13px;color:#666">Loading...</p>',a==null||a.appendChild(e),p();try{let i=((await(await fetch(`${k}/api/booking/availability?tid=${v}`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({})})).json()).slots||[]).slice(0,8);if(i.length===0){e.innerHTML='<h4>Available Times</h4><p style="color:#666;font-size:13px">No slots available right now. Please contact us directly.</p>';return}let r=document.createElement("div");r.className="bp-chat-slot-list",i.forEach(c=>{let s=document.createElement("button");s.className="bp-chat-slot-btn";let x=new Date(c.startTime);s.textContent=x.toLocaleString("en-US",{weekday:"short",month:"short",day:"numeric",hour:"numeric",minute:"2-digit",hour12:!0}),s.addEventListener("click",()=>V(c,e)),r.appendChild(s)}),e.innerHTML="<h4>Pick a time</h4>",e.appendChild(r),p()}catch(n){e.innerHTML='<div class="bp-chat-error-msg">Could not load available times. Please contact us directly.</div>'}}function V(e,n){var t;n.innerHTML=`
      <h4>Confirm your booking</h4>
      <p style="font-size:13px; color:#444">${new Date(e.startTime).toLocaleString("en-US",{weekday:"long",month:"long",day:"numeric",hour:"numeric",minute:"2-digit",hour12:!0})}</p>
      <input class="bp-chat-lead-field" name="name" placeholder="Your name *" />
      <input class="bp-chat-lead-field" type="email" name="email" placeholder="Email address *" />
      <input class="bp-chat-lead-field" name="notes" placeholder="Anything we should know? (optional)" />
      <button class="bp-chat-lead-submit">Confirm Booking</button>
    `,(t=n.querySelector(".bp-chat-lead-submit"))==null||t.addEventListener("click",async()=>{var x,y,m;let i=((x=n.querySelector('[name="name"]'))==null?void 0:x.value)||"",r=((y=n.querySelector('[name="email"]'))==null?void 0:y.value)||"",c=((m=n.querySelector('[name="notes"]'))==null?void 0:m.value)||"";if(!i||!r){alert("Name and email are required.");return}let s=n.querySelector(".bp-chat-lead-submit");s.disabled=!0,s.textContent="Booking...";try{let b=await fetch(`${k}/api/booking/create?tid=${v}`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({slotStartTime:e.startTime,attendeeName:i,attendeeEmail:r,notes:c})});if(b.ok){let f=await b.json();n.innerHTML=`
            <div class="bp-chat-confirm-msg">
              \u2713 <strong>Booking confirmed!</strong><br/>
              ${new Date(f.startTime||e.startTime).toLocaleString("en-US",{weekday:"long",month:"long",day:"numeric",hour:"numeric",minute:"2-digit",hour12:!0})}
              ${f.meetingUrl?`<br/><a href="${f.meetingUrl}" target="_blank" style="color:#2e7d32">Join meeting link</a>`:""}
              ${f.confirmUrl?`<br/><a href="${f.confirmUrl}" target="_blank" style="color:#2e7d32; font-size:12px">View booking details</a>`:""}
            </div>`}else s.disabled=!1,s.textContent="Try again"}catch(b){s.disabled=!1,s.textContent="Try again"}p()}),p()}function K(e){u=document.createElement("button"),u.className=`bp-chat-bubble bp-chat-bubble--${e.position}`,u.setAttribute("aria-label",`Chat with ${e.botName}`),u.innerHTML='<svg viewBox="0 0 24 24"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/></svg>',u.addEventListener("click",X),document.body.appendChild(u)}function G(e){var n;l=document.createElement("div"),l.className=`bp-chat-panel bp-chat-panel--${e.position} bp-chat-panel--hidden`,l.innerHTML=`
      <div class="bp-chat-header">
        <div class="bp-chat-header-avatar">
          ${e.avatarUrl?`<img src="${e.avatarUrl}" alt="${e.botName}" />`:e.botName.charAt(0)}
        </div>
        <div class="bp-chat-header-info">
          <div class="bp-chat-header-name">${e.botName}</div>
          <div class="bp-chat-header-status">Online \xB7 Usually replies instantly</div>
        </div>
        <button class="bp-chat-header-close" aria-label="Close chat">
          <svg viewBox="0 0 24 24"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
        </button>
      </div>
      <div class="bp-chat-messages"></div>
      <div class="bp-chat-input-area">
        <textarea class="bp-chat-input" placeholder="${e.placeholderText||"Type a message..."}" rows="1"></textarea>
        <button class="bp-chat-send-btn" aria-label="Send">
          <svg viewBox="0 0 24 24"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>
        </button>
      </div>
      <div class="bp-chat-powered"><a href="https://blueprintautomation.tech" target="_blank" rel="noopener">Powered by Blueprint Chat</a></div>
    `,a=l.querySelector(".bp-chat-messages"),d=l.querySelector(".bp-chat-input"),S=l.querySelector(".bp-chat-send-btn"),(n=l.querySelector(".bp-chat-header-close"))==null||n.addEventListener("click",P),S==null||S.addEventListener("click",z),d==null||d.addEventListener("keydown",t=>{t.key==="Enter"&&!t.shiftKey&&(t.preventDefault(),z())}),d==null||d.addEventListener("input",()=>Q(d)),document.body.appendChild(l)}function Q(e){e.style.height="auto",e.style.height=Math.min(e.scrollHeight,100)+"px"}function z(){if(!d||!S||E==="typing")return;let e=d.value.trim();e&&(d.value="",d.style.height="auto",W(e))}function X(){l&&(l.classList.contains("bp-chat-panel--hidden")?Z():P())}function Z(){l==null||l.classList.remove("bp-chat-panel--hidden"),u==null||u.classList.remove("bp-chat-bubble--pulse"),L("open"),M.length===0&&o&&C("assistant",o.greeting),d==null||d.focus()}function P(){l==null||l.classList.add("bp-chat-panel--hidden"),L("closed"),T==null||T.abort()}async function B(){try{o=await U()}catch(e){console.error("[Blueprint Chat] Failed to load config:",e);return}O(o.primaryColor),K(o),G(o)}document.readyState==="loading"?document.addEventListener("DOMContentLoaded",B):B()})();})();

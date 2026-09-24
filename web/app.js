(() => {
  'use strict';
  const $ = (s) => document.querySelector(s);
  const $$ = (s) => document.querySelectorAll(s);
  const form = $('#booking-form');
  const status = $('#form-status');
  const btn = $('#book-btn');
  let cfg = { villaName: 'Villa Les Oliviers', pricePerNight: 350, currency: 'EUR', maxGuests: 8, minNights: 2 };

  const money = (n) => new Intl.NumberFormat('en-GB', {
    style: 'currency', currency: cfg.currency, maximumFractionDigits: 0,
  }).format(n);
  const iso = (d) => d.toISOString().slice(0, 10);
  const nights = (a, b) => (a && b ? Math.round((Date.parse(b) - Date.parse(a)) / 86400000) : 0);
  const nice = (s) => new Date(s + 'T00:00:00Z').toLocaleDateString('en-GB', {
    day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC',
  });

  function applyConfig() {
    $$('.js-villa-name').forEach((el) => { el.textContent = cfg.villaName; });
    $$('.js-max-guests').forEach((el) => { el.textContent = cfg.maxGuests; });
    $$('.js-min-nights').forEach((el) => { el.textContent = cfg.minNights; });
    $$('.js-price').forEach((el) => { el.textContent = money(cfg.pricePerNight); });
    const sel = $('#guests');
    sel.innerHTML = '';
    for (let i = 1; i <= cfg.maxGuests; i++) sel.add(new Option(String(i), String(i), i === 2, i === 2));
    updateEstimate();
  }

  async function loadConfig() {
    try {
      const r = await fetch('/api/config');
      if (r.ok) cfg = { ...cfg, ...(await r.json()) };
    } catch (_) { /* keep defaults */ }
    applyConfig();
  }

  async function loadAvailability() {
    const list = $('#booked-list');
    try {
      const r = await fetch('/api/availability');
      const data = await r.json();
      list.innerHTML = '';
      if (!data.booked || data.booked.length === 0) {
        list.innerHTML = '<li>No bookings yet: all dates are open!</li>';
        return;
      }
      data.booked.forEach((b) => {
        const li = document.createElement('li');
        li.textContent = `${nice(b.checkIn)} → ${nice(b.checkOut)}`;
        list.appendChild(li);
      });
    } catch (_) {
      list.innerHTML = '<li>Availability could not be loaded.</li>';
    }
  }

  function updateEstimate() {
    const n = nights(form.checkIn.value, form.checkOut.value);
    const el = $('#estimate');
    if (n <= 0) { el.textContent = `${money(cfg.pricePerNight)} per night`; return; }
    const warn = n < cfg.minNights ? ` (minimum ${cfg.minNights} nights)` : '';
    el.textContent = `${n} night${n > 1 ? 's' : ''} × ${money(cfg.pricePerNight)} = ${money(n * cfg.pricePerNight)}${warn}`;
  }

  function setStatus(msg, ok) {
    status.textContent = msg;
    status.className = 'status ' + (ok ? 'ok' : 'err');
  }

  form.checkIn.addEventListener('change', () => {
    if (form.checkIn.value) {
      const next = new Date(form.checkIn.value);
      next.setUTCDate(next.getUTCDate() + cfg.minNights);
      form.checkOut.min = iso(next);
      if (!form.checkOut.value || form.checkOut.value < form.checkOut.min) form.checkOut.value = iso(next);
    }
    updateEstimate();
  });
  form.checkOut.addEventListener('change', updateEstimate);

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!form.reportValidity()) return;
    const payload = Object.fromEntries(new FormData(form).entries());
    payload.guests = Number(payload.guests);
    btn.disabled = true;
    setStatus('Sending your booking…', true);
    try {
      const r = await fetch('/api/bookings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const data = await r.json().catch(() => ({}));
      if (!r.ok) throw new Error(data.error || 'Something went wrong, please try again.');
      setStatus(`Thank you! Booking #${data.id} received (${data.nights} nights, ${money(data.total)}). A confirmation email is on its way.`, true);
      form.reset();
      applyConfig();
      loadAvailability();
    } catch (err) {
      setStatus(err.message, false);
    } finally {
      btn.disabled = false;
    }
  });

  const today = iso(new Date());
  form.checkIn.min = today;
  form.checkOut.min = today;
  $('#year').textContent = new Date().getFullYear();
  loadConfig();
  loadAvailability();
})();

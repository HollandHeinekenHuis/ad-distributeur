# ad-distributeur: Villa rental website

Single-page villa rental site (HTML/CSS/JS) served by a Go backend, with
bookings stored in a CloudNativePG PostgreSQL cluster.

## Layout

- `web/`: the single-page site (embedded into the Go binary)
- `web/images/`: put your photos here (see `web/images/README.md` for names)
- `*.go`: HTTP server, booking API, validation, email
- `k8s/`: CNPG operator + `villa-db` cluster, app Deployment, Service, HTTPRoute

## API

- `GET /api/config`: villa name, nightly price, guests, minimum nights
- `GET /api/availability`: upcoming booked date ranges
- `POST /api/bookings`: `{name, email, phone, guests, checkIn, checkOut, message}`
  → 201, or 409 if the dates overlap an existing booking

## Settings (env vars in `k8s/deployment.yaml`)

`VILLA_NAME`, `PRICE_PER_NIGHT`, `CURRENCY`, `MAX_GUESTS`, `MIN_NIGHTS`.

## Email

Every booking triggers a confirmation email to the guest and a notification
to `OWNER_EMAIL`. While `SMTP_HOST` is empty, emails are only written to the
app log. To enable sending, set `SMTP_HOST`, `SMTP_PORT`, `MAIL_FROM`,
`OWNER_EMAIL`, and provide `SMTP_USER` / `SMTP_PASSWORD` via a Secret named
`smtp-credentials` (keys `username`, `password`).

# ad-distributeur: Villa rental website

A direct-booking website for **Villa Les Oliviers**, a restored 18th-century
stone farmhouse with a private pool in the hills of Provence, France.

Guests can browse the villa, its amenities and location on a single page, see
which dates are already taken, and send a booking request without going
through a third-party rental platform. The owner gets an email for every new
booking, and so does the guest.

## What it does

- **Showcase page:** hero, about the villa (bedrooms, bathrooms, capacity,
  nightly price), photo gallery, amenities and location, all on one page.
- **Live availability:** upcoming booked periods are loaded from the database
  and listed next to the booking form.
- **Booking form:** the guest picks check-in/check-out dates and the number of
  guests and enters their contact details. The page shows the number of nights
  and the total price before they submit.
- **Server-side validation:** the server checks dates, the minimum stay,
  maximum guests and contact fields. It also refuses overlapping bookings.
  Bookings run one at a time (a PostgreSQL advisory lock), so two guests can
  never reserve the same nights.
- **Email notifications:** a confirmation goes to the guest and a notification
  to the owner. Until SMTP is set up, the emails are written to the app log.
- **Configurable without code changes:** villa name, price, currency, capacity
  and minimum stay come from environment variables.

## How it's built

- **Frontend:** plain HTML, CSS and JavaScript in `web/`. It is embedded into
  the Go binary with `go:embed`, so the app ships as one small container.
- **Backend:** a Go server using only the standard library `net/http`, plus
  `pgx` for PostgreSQL. It serves the site and a small JSON API.
- **Database:** PostgreSQL managed by the CloudNativePG operator (a
  `villa-db` cluster). The server creates its schema at startup and retries
  until the database is reachable. Until then, `/readyz` reports not ready.
- **Deployment:** Kubernetes manifests in `k8s/`, built and deployed by the
  GitHub Actions workflow. The container runs as non-root with a read-only
  root filesystem.

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
- `GET /healthz`: liveness; `GET /readyz`: readiness (database reachable)

## Settings (env vars in `k8s/deployment.yaml`)

`VILLA_NAME`, `PRICE_PER_NIGHT`, `CURRENCY`, `MAX_GUESTS`, `MIN_NIGHTS`.

## Email

Every booking triggers a confirmation email to the guest and a notification
to `OWNER_EMAIL`. While `SMTP_HOST` is empty, emails are only written to the
app log. To enable sending, set `SMTP_HOST`, `SMTP_PORT`, `MAIL_FROM`,
`OWNER_EMAIL`, and provide `SMTP_USER` / `SMTP_PASSWORD` via a Secret named
`smtp-credentials` (keys `username`, `password`).

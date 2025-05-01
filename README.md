# Beam

Beam is a modern, extensible e-commerce platform designed to host multiple storefronts across distinct domains, built on a robust Go Gin monolith using the recommended repository and service pattern for large-scale maintainability.

## Key Features

- **Multi-Store Hosting**: Serve multiple stores from a single instance, each on its own domain.
- **Flexible Data Storage**:
  - Small data (e.g., country/state lists) in RWMutex memory stores.
  - Medium-sized, frequently accessed data (e.g., product and auth info) in Redis.
  - Larger, less sensitive data (e.g., carts, save-for-later lists) in Postgres.
  - Very large data (e.g., orders) in MongoDB.
- **Built-in Auth System**: Includes 2FA, Turnstile integration, rate limiting (per IP/device/account), and strong encryption.
- **Global Cart Settings**: Smart conflict resolution for collaborative cart editing within the same account.
- **Stripe Integration**: Fast and reliable webhook handling with remediation for failed payments.
- **Filterable URLs with HTMX**: Product filtering modifies the URL directly, enabling shareable/filter-persistent links without full reloads.
- **Detailed User Interaction Logging**: Logs user interaction down to the function level for superior debugging and analytics.

## Tech Stack

- **Backend**: Go (Gin), Redis, Postgres, MongoDB
- **Frontend** (in progress): Alpine.js, HTMX, Bootstrap
- **Hosting**: Monolithic architecture capable of supporting scale and modular growth

## Roadmap

The backend is nearing completion, while the frontend is under active development. The goal is to deliver a mobile-first, performance-optimized interface suitable for low-powered devices by offloading most of the heavy work to the backend.

This is a long-term project with development measured in months and years. Progress is ongoing.

## Contact

For questions, ideas, or collaboration, please reach out via the [Contact Me form](https://kopatsis.com/#contact).

# ADR 003: Use Supabase for Data Storage

## Status
Accepted

## Context
We needed a reliable, scalable database solution that could handle album and lyrics data with good performance and easy integration.

## Decision
We chose Supabase because:
- PostgreSQL-based with full SQL capabilities
- Built-in authentication and authorization
- Real-time capabilities
- RESTful API out of the box
- Open source and self-hostable
- Good Go client support
- Free tier suitable for development

## Consequences
Positive:
- Robust data storage and querying capabilities
- Built-in security features
- Scalable solution
- Good developer experience
- Cost-effective for small to medium scale

Negative:
- Additional dependency on external service
- Learning curve for PostgreSQL if not familiar
- Potential vendor lock-in
- Requires internet connection for cloud version 
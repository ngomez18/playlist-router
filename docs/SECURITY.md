# Security Architecture

This document outlines the security measures implemented in PlaylistRouter, specifically focusing on the protection of sensitive user data and authentication tokens.

## Data Encryption at Rest

We strictly adhere to the principle of least privilege and defense in depth. Sensitive third-party credentials, specifically Spotify OAuth tokens, are **never** stored in plain text in our database.

### Encryption Standard

We use **AES-256-GCM** (Advanced Encryption Standard with Galois/Counter Mode) for all data encryption.

-   **Algorithm**: AES (Advanced Encryption Standard)
-   **Key Size**: 256-bit (32 bytes)
-   **Mode**: GCM (Galois/Counter Mode) - Provides both confidentiality and data integrity (authenticated encryption).
-   **Nonce**: Unique, randomly generated 12-byte nonce for every encryption operation.

### Key Management

The encryption key is managed via environment variables and is never hardcoded in the application source code.

-   **Variable**: `ENCRYPTION_KEY`
-   **Requirement**: Must be a 32-byte string.
-   **Storage**: In production, this should be injected via a secure secret manager (e.g., AWS Secrets Manager, Google Secret Manager, or Kubernetes Secrets).

### Generating a Secure Key

You can generate a secure 32-character key using `openssl` in your terminal. This command generates 16 random bytes and converts them to a 32-character hexadecimal string, which satisfies the application's length requirement:

```bash
openssl rand -hex 16
```

Example output: `a1b2c3d4e5f67890123456789abcdef0`

Copy this value and set it as your `ENCRYPTION_KEY` in the `.env` file.

### Encryption Process

When sensitive data (like a Spotify Access Token) needs to be stored:

1.  **Input**: The plaintext string (e.g., the token).
2.  **Nonce Generation**: A cryptographically secure random 12-byte nonce is generated.
3.  **Encryption**: The plaintext is encrypted using AES-GCM with the `ENCRYPTION_KEY` and the generated nonce.
4.  **Sealing**: The GCM "seal" operation appends the authentication tag to the ciphertext to ensure integrity.
5.  **Encoding**: The combined `[nonce + ciphertext + tag]` is Base64 encoded for safe storage as a text string in the database.

### Decryption Process

When the application needs to use a token:

1.  **Retrieval**: The Base64 string is retrieved from the database.
2.  **Decoding**: The string is Base64 decoded back into raw bytes.
3.  **Extraction**: The nonce is extracted from the first 12 bytes of the data.
4.  **Decryption**: The remaining bytes (ciphertext + tag) are decrypted using AES-GCM with the `ENCRYPTION_KEY` and the extracted nonce.
5.  **Verification**: GCM automatically verifies the authentication tag. If the data has been tampered with, decryption fails.
6.  **Output**: The original plaintext string is returned for use in memory.

## Authentication

-   **User Auth**: Handled via PocketBase's built-in JWT system.
-   **Session Management**: Stateless JWTs with expiration.
-   **Spotify Auth**: OAuth 2.0 flow. Access and Refresh tokens are encrypted as described above.

## Security Best Practices

-   **HTTPS**: All traffic must be encrypted in transit (TLS/SSL).
-   **Logs**: Sensitive data (tokens, keys, passwords) must **never** be logged.
